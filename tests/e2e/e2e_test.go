//go:build e2e

package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	steprpcv1 "github.com/albertocavalcante/jenkins-rpc/contracts/gen/go/proto/steprpc/v1"
	jenkinsrpc "github.com/albertocavalcante/jenkins-rpc/go-client"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	jenkinsURL     = "http://localhost:8080"
	startupTimeout = 120 * time.Second
)

func TestMain(m *testing.M) {
	manageDocker := os.Getenv("E2E_SKIP_DOCKER") == ""

	if manageDocker {
		if err := dockerComposeUp(); err != nil {
			fmt.Fprintf(os.Stderr, "docker compose up: %v\n", err)
			dockerComposeDown() //nolint:errcheck // best-effort cleanup
			os.Exit(1)
		}
	}

	if err := waitForJenkins(jenkinsURL, startupTimeout); err != nil {
		fmt.Fprintf(os.Stderr, "jenkins startup: %v\n", err)
		if manageDocker {
			dockerComposeDown() //nolint:errcheck // best-effort cleanup
		}
		os.Exit(1)
	}

	code := m.Run()

	if manageDocker {
		dockerComposeDown() //nolint:errcheck // best-effort cleanup
	}
	os.Exit(code)
}

func TestHealth(t *testing.T) {
	resp, err := http.Get(jenkinsURL + "/step-rpc/v1/")
	if err != nil {
		t.Fatalf("GET /step-rpc/v1/: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	health := &steprpcv1.HealthResponse{}
	if err := protojson.Unmarshal(body, health); err != nil {
		t.Fatalf("unmarshal health response: %v", err)
	}

	if health.GetStatus() != "ok" {
		t.Errorf("status: got %q, want %q", health.GetStatus(), "ok")
	}
	if health.GetApiVersion() != "v1" {
		t.Errorf("apiVersion: got %q, want %q", health.GetApiVersion(), "v1")
	}
	if health.GetService() != "jenkins-step-rpc-plugin" {
		t.Errorf("service: got %q, want %q", health.GetService(), "jenkins-step-rpc-plugin")
	}
}

func TestCatalog(t *testing.T) {
	client, err := jenkinsrpc.New(jenkinsURL, "", nil)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	catalog, err := client.GetCatalog(ctx)
	if err != nil {
		t.Fatalf("get catalog: %v", err)
	}

	if len(catalog.GetOperations()) == 0 {
		t.Fatal("catalog has no operations")
	}

	// archiveArtifacts must be present with DIRECT execution mode.
	found := false
	for _, op := range catalog.GetOperations() {
		if op.GetName() == "archiveArtifacts" {
			found = true
			if op.GetExecutionMode() != steprpcv1.OperationExecutionMode_OPERATION_EXECUTION_MODE_DIRECT {
				t.Errorf("archiveArtifacts execution mode: got %v, want DIRECT", op.GetExecutionMode())
			}
			break
		}
	}
	if !found {
		t.Error("archiveArtifacts not found in catalog")
	}

	// DirectOperations helper should include archiveArtifacts.
	directOps := jenkinsrpc.DirectOperations(catalog)
	foundInHelper := false
	for _, name := range directOps {
		if name == "archiveArtifacts" {
			foundInHelper = true
			break
		}
	}
	if !foundInHelper {
		t.Errorf("DirectOperations() did not include archiveArtifacts: %v", directOps)
	}
}

func TestDirectInvoke(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	jobName := "e2e-archive-test"

	// 1. Create a freestyle job that writes artifact.txt.
	jobXML := `<?xml version='1.1' encoding='UTF-8'?>
<project>
  <builders>
    <hudson.tasks.Shell>
      <command>echo "hello from e2e" > artifact.txt</command>
    </hudson.tasks.Shell>
  </builders>
</project>`

	if err := createJob(ctx, jenkinsURL, jobName, jobXML); err != nil {
		t.Fatalf("create job: %v", err)
	}

	// 2. Trigger build and wait for completion.
	if err := triggerBuild(ctx, jenkinsURL, jobName); err != nil {
		t.Fatalf("trigger build: %v", err)
	}
	if err := waitForBuild(ctx, jenkinsURL, jobName, 1); err != nil {
		t.Fatalf("wait for build: %v", err)
	}

	// 3. Invoke archiveArtifacts via RPC.
	client, err := jenkinsrpc.New(jenkinsURL, "", nil)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	args, err := structpb.NewStruct(map[string]any{
		"artifacts": "artifact.txt",
		"runContext": map[string]any{
			"jobFullName": jobName,
			"buildNumber": 1,
			"nodeName":    "built-in",
			"workspace":   "/var/jenkins_home/workspace/" + jobName,
		},
	})
	if err != nil {
		t.Fatalf("build args struct: %v", err)
	}

	invokeResp, err := client.Invoke(ctx, &steprpcv1.InvokeRequest{
		RequestId: "e2e-req-1",
		Operation: "archiveArtifacts",
		Args:      args,
	})
	if err != nil {
		t.Fatalf("invoke: %v", err)
	}

	// 4. Assert state is "succeeded".
	if invokeResp.GetState() != "succeeded" {
		t.Errorf("invoke state: got %q, want %q", invokeResp.GetState(), "succeeded")
	}

	// 5. Verify artifact exists via Jenkins REST API.
	artifacts, err := getBuildArtifacts(ctx, jenkinsURL, jobName, 1)
	if err != nil {
		t.Fatalf("get build artifacts: %v", err)
	}

	foundArtifact := false
	for _, a := range artifacts {
		if a == "artifact.txt" {
			foundArtifact = true
			break
		}
	}
	if !foundArtifact {
		t.Errorf("artifact.txt not found in build artifacts: %v", artifacts)
	}
}

// getBuildArtifacts returns artifact file names for a build.
func getBuildArtifacts(ctx context.Context, baseURL, jobName string, buildNum int) ([]string, error) {
	url := fmt.Sprintf("%s/job/%s/%d/api/json?tree=artifacts[fileName]", baseURL, jobName, buildNum)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get artifacts: status=%d body=%s", resp.StatusCode, body)
	}

	var result struct {
		Artifacts []struct {
			FileName string `json:"fileName"`
		} `json:"artifacts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode artifacts: %w", err)
	}

	var names []string
	for _, a := range result.Artifacts {
		names = append(names, a.FileName)
	}
	return names, nil
}
