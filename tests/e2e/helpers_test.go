//go:build e2e

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

const composeDir = "docker"

// dockerComposeUp builds and starts the Jenkins container.
func dockerComposeUp() error {
	cmd := exec.Command("docker", "compose", "up", "--build", "--wait")
	cmd.Dir = composeDir
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// dockerComposeDown tears down the Jenkins container and removes volumes.
func dockerComposeDown() error {
	cmd := exec.Command("docker", "compose", "down", "-v")
	cmd.Dir = composeDir
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// waitForJenkins polls the health endpoint until it responds 200 or timeout expires.
func waitForJenkins(baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/step-rpc/v1/")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("jenkins did not become healthy within %v", timeout)
}

// createJob creates a freestyle job via Jenkins REST API.
func createJob(ctx context.Context, baseURL, name, configXML string) error {
	url := fmt.Sprintf("%s/createItem?name=%s", baseURL, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(configXML))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/xml")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create job: status=%d body=%s", resp.StatusCode, body)
	}
	return nil
}

// triggerBuild starts a build for the given job.
func triggerBuild(ctx context.Context, baseURL, jobName string) error {
	url := fmt.Sprintf("%s/job/%s/build", baseURL, jobName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 201 Created is the expected response for build trigger.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("trigger build: status=%d body=%s", resp.StatusCode, body)
	}
	return nil
}

// waitForBuild polls until the specified build completes or the context expires.
func waitForBuild(ctx context.Context, baseURL, jobName string, buildNum int) error {
	url := fmt.Sprintf("%s/job/%s/%d/api/json?tree=building,result", baseURL, jobName, buildNum)

	for {
		// First check if the build exists yet (it may be queued).
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			select {
			case <-ctx.Done():
				return fmt.Errorf("wait for build: %w", ctx.Err())
			case <-time.After(1 * time.Second):
				continue
			}
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return fmt.Errorf("wait for build: status=%d body=%s", resp.StatusCode, body)
		}

		var result struct {
			Building bool    `json:"building"`
			Result   *string `json:"result"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return fmt.Errorf("decode build status: %w", err)
		}
		resp.Body.Close()

		if !result.Building && result.Result != nil {
			if *result.Result != "SUCCESS" {
				return fmt.Errorf("build finished with result: %s", *result.Result)
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait for build: %w", ctx.Err())
		case <-time.After(1 * time.Second):
		}
	}
}
