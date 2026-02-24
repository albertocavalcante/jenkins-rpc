package rpcclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	steprpcv1 "github.com/albertocavalcante/jenkins-rpc/contracts/gen/go/proto/steprpc/v1"
)

const operationJunit = "junit"

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name)) //nolint:gosec // test fixture paths are not user-controlled
	if err != nil {
		t.Fatalf("load fixture %s: %v", name, err)
	}
	return data
}

func fixtureServer(t *testing.T, status int, fixture string) *httptest.Server {
	t.Helper()
	body := loadFixture(t, fixture)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(ts.Close)
	return ts
}

func TestFixture_InvokeSuccess(t *testing.T) {
	t.Parallel()
	ts := fixtureServer(t, http.StatusOK, "invoke_success.json")

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := c.Invoke(context.Background(), &steprpcv1.InvokeRequest{
		RequestId: "r-1",
		Operation: "test",
	})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if resp.GetRunId() != "run-1" {
		t.Fatalf("runId = %s, want run-1", resp.GetRunId())
	}
	if resp.GetState() != "queued" {
		t.Fatalf("state = %s, want queued", resp.GetState())
	}
}

func TestFixture_InvokeError(t *testing.T) {
	t.Parallel()
	ts := fixtureServer(t, http.StatusBadRequest, "invoke_error.json")

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.Invoke(context.Background(), &steprpcv1.InvokeRequest{
		RequestId: "r-1",
		Operation: "unknown",
	})
	assertHTTPError(t, err, http.StatusBadRequest, "operation_not_allowed")
	if got := CategoryOf(err); got != CategoryBadRequest {
		t.Fatalf("CategoryOf() = %v, want BadRequest", got)
	}
}

func TestFixture_RunStatusSucceeded(t *testing.T) {
	t.Parallel()
	ts := fixtureServer(t, http.StatusOK, "run_status_succeeded.json")

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := c.GetRunStatus(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("GetRunStatus() error = %v", err)
	}
	if resp.GetState() != stateSucceeded {
		t.Fatalf("state = %s, want %s", resp.GetState(), stateSucceeded)
	}
}

func TestFixture_RunStatusFailed(t *testing.T) {
	t.Parallel()
	ts := fixtureServer(t, http.StatusOK, "run_status_failed.json")

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := c.GetRunStatus(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("GetRunStatus() error = %v", err)
	}
	if resp.GetState() != "failed" {
		t.Fatalf("state = %s, want failed", resp.GetState())
	}
	if resp.GetError().GetCode() != "operation_failed" {
		t.Fatalf("error code = %s, want operation_failed", resp.GetError().GetCode())
	}
}

func TestFixture_Catalog(t *testing.T) {
	t.Parallel()
	ts := fixtureServer(t, http.StatusOK, "catalog.json")

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	catalog, err := c.GetCatalog(context.Background())
	if err != nil {
		t.Fatalf("GetCatalog() error = %v", err)
	}
	if len(catalog.GetOperations()) != 2 {
		t.Fatalf("operations len = %d, want 2", len(catalog.GetOperations()))
	}
	if got := DirectOperations(catalog); len(got) != 1 || got[0] != "archiveArtifacts" {
		t.Fatalf("DirectOperations = %v", got)
	}
}

func TestFixture_BridgePending(t *testing.T) {
	t.Parallel()
	ts := fixtureServer(t, http.StatusOK, "bridge_pending.json")

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := c.GetBridgePending(context.Background(), "job/demo#1")
	if err != nil {
		t.Fatalf("GetBridgePending() error = %v", err)
	}
	if resp.GetOperation() != operationJunit {
		t.Fatalf("operation = %s, want %s", resp.GetOperation(), operationJunit)
	}
}

func TestFixture_BridgeComplete(t *testing.T) {
	t.Parallel()
	ts := fixtureServer(t, http.StatusOK, "bridge_complete.json")

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := c.CompleteBridgeRequest(context.Background(), &steprpcv1.BridgeCompleteRequest{
		RunId: "rpc-1",
		State: "succeeded",
	})
	if err != nil {
		t.Fatalf("CompleteBridgeRequest() error = %v", err)
	}
	if resp.GetState() != stateSucceeded {
		t.Fatalf("state = %s, want %s", resp.GetState(), stateSucceeded)
	}
}

func TestFixture_ServerError(t *testing.T) {
	t.Parallel()
	ts := fixtureServer(t, http.StatusInternalServerError, "server_error.json")

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.GetCatalog(context.Background())
	assertHTTPError(t, err, http.StatusInternalServerError, "internal_error")
	if got := CategoryOf(err); got != CategoryServerError {
		t.Fatalf("CategoryOf() = %v, want ServerError", got)
	}
}

func TestFixture_RateLimited(t *testing.T) {
	t.Parallel()
	ts := fixtureServer(t, http.StatusTooManyRequests, "rate_limited.json")

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.GetCatalog(context.Background())
	assertHTTPError(t, err, http.StatusTooManyRequests, "rate_limited")
	if got := CategoryOf(err); got != CategoryRateLimited {
		t.Fatalf("CategoryOf() = %v, want RateLimited", got)
	}
}
