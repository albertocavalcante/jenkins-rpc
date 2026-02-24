package rpcclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	steprpcv1 "github.com/albertocavalcante/jenkins-rpc/contracts/gen/go/proto/steprpc/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

const stateSucceeded = "succeeded"

func TestInvoke(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/step-rpc/v1/invoke" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"r-1","runId":"job#1","state":"queued"}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	args, err := structpb.NewStruct(map[string]any{
		"artifacts": "build/*.jar",
	})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := c.Invoke(context.Background(), &steprpcv1.InvokeRequest{
		RequestId: "r-1",
		Operation: "archiveArtifacts",
		Args:      args,
	})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if resp.GetRunId() != "job#1" {
		t.Fatalf("runID = %s, want job#1", resp.GetRunId())
	}
}

func TestInvoke_ErrorPayload(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":{"code":"operation_not_allowed","message":"denied"}}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.Invoke(context.Background(), &steprpcv1.InvokeRequest{
		RequestId: "r-1",
		Operation: "unknown",
	})
	if err == nil {
		t.Fatalf("Invoke() error = nil, want non-nil")
	}

	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error type = %T, want *HTTPError", err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("statusCode = %d, want 400", httpErr.StatusCode)
	}
	if httpErr.ProtoError == nil || httpErr.ProtoError.GetCode() != "operation_not_allowed" {
		t.Fatalf("proto error code = %v", httpErr.ProtoError)
	}
}

func TestGetRunStatus(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		prefix := "/step-rpc/v1/runs/"
		if !strings.HasPrefix(r.URL.Path, prefix) {
			t.Fatalf("path = %s", r.URL.Path)
		}
		runID, err := url.PathUnescape(strings.TrimPrefix(r.URL.Path, prefix))
		if err != nil {
			t.Fatalf("PathUnescape() error = %v", err)
		}
		if runID != "job#1" {
			t.Fatalf("runID = %s, want job#1", runID)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"r-1","runId":"job#1","operation":"archiveArtifacts","state":"succeeded"}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := c.GetRunStatus(context.Background(), "job#1")
	if err != nil {
		t.Fatalf("GetRunStatus() error = %v", err)
	}
	if resp.GetState() != stateSucceeded {
		t.Fatalf("state = %s, want %s", resp.GetState(), stateSucceeded)
	}
}

func TestGetRunStatus_ErrorPayload(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":{"code":"run_not_found","message":"missing"}}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.GetRunStatus(context.Background(), "missing")
	if err == nil {
		t.Fatalf("GetRunStatus() error = nil, want non-nil")
	}

	assertHTTPError(t, err, http.StatusNotFound, "run_not_found")
}

func TestGetBridgePending(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/step-rpc/v1/bridge/pending" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("runExternalizableId") != "job/demo#1" {
			t.Fatalf("query runExternalizableId = %s", r.URL.Query().Get("runExternalizableId"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"req-1","runId":"rpc-1","operation":"junit","args":{"testResults":"**/*.xml"},"targetRunExternalizableId":"job/demo#1"}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := c.GetBridgePending(context.Background(), "job/demo#1")
	if err != nil {
		t.Fatalf("GetBridgePending() error = %v", err)
	}
	if resp.GetRunId() != "rpc-1" {
		t.Fatalf("runID = %s, want rpc-1", resp.GetRunId())
	}
	if resp.GetOperation() != "junit" {
		t.Fatalf("operation = %s, want junit", resp.GetOperation())
	}
}

func TestGetBridgePending_ErrorPayload(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":{"code":"no_pending_request","message":"none"}}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.GetBridgePending(context.Background(), "job/demo#1")
	if err == nil {
		t.Fatalf("GetBridgePending() error = nil, want non-nil")
	}

	assertHTTPError(t, err, http.StatusNotFound, "no_pending_request")
}

func TestCompleteBridgeRequest(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/step-rpc/v1/bridge/complete" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("content-type = %s, want application/json", ct)
		}

		var in steprpcv1.BridgeCompleteRequest
		body := mustReadAll(t, r.Body)
		if err := protojson.Unmarshal(body, &in); err != nil {
			t.Fatalf("Unmarshal() error = %v", err)
		}
		if in.GetRunId() != "rpc-1" || in.GetState() != stateSucceeded {
			t.Fatalf("runID/state = %s/%s, want rpc-1/%s", in.GetRunId(), in.GetState(), stateSucceeded)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"req-1","runId":"rpc-1","state":"succeeded"}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := c.CompleteBridgeRequest(context.Background(), &steprpcv1.BridgeCompleteRequest{
		RunId: "rpc-1",
		State: stateSucceeded,
	})
	if err != nil {
		t.Fatalf("CompleteBridgeRequest() error = %v", err)
	}
	if resp.GetState() != stateSucceeded {
		t.Fatalf("state = %s, want %s", resp.GetState(), stateSucceeded)
	}
}

func TestCompleteBridgeRequest_ErrorPayload(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":{"code":"run_not_found","message":"missing"}}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.CompleteBridgeRequest(context.Background(), &steprpcv1.BridgeCompleteRequest{
		RunId: "rpc-1",
		State: "failed",
		Error: &steprpcv1.Error{
			Code:    "operation_failed",
			Message: "boom",
		},
	})
	if err == nil {
		t.Fatalf("CompleteBridgeRequest() error = nil, want non-nil")
	}

	assertHTTPError(t, err, http.StatusNotFound, "run_not_found")
}

func TestGetCatalog(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/step-rpc/v1/catalog" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"operations":[{"name":"archiveArtifacts","description":"Archive (direct)","executionMode":"OPERATION_EXECUTION_MODE_DIRECT"},{"name":"junit","description":"Publish (CPS context required)","executionMode":"OPERATION_EXECUTION_MODE_CPS_BRIDGE_REQUIRED"}]}`))
	}))
	defer ts.Close()

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

	direct := DirectOperations(catalog)
	if len(direct) != 1 || direct[0] != "archiveArtifacts" {
		t.Fatalf("direct operations = %+v", direct)
	}

	cps := CPSBridgeOperations(catalog)
	if len(cps) != 1 || cps[0] != "junit" {
		t.Fatalf("cps operations = %+v", cps)
	}
}

func TestGetCatalog_ErrorPayload(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"error":{"code":"unauthorized","message":"denied"}}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.GetCatalog(context.Background())
	if err == nil {
		t.Fatalf("GetCatalog() error = nil, want non-nil")
	}
	assertHTTPError(t, err, http.StatusUnauthorized, "unauthorized")
}

func TestWaitRunTerminal(t *testing.T) {
	t.Parallel()

	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		if call < 3 {
			_, _ = w.Write([]byte(`{"requestId":"r-1","runId":"job#1","operation":"archiveArtifacts","state":"running"}`))
			return
		}
		_, _ = w.Write([]byte(`{"requestId":"r-1","runId":"job#1","operation":"archiveArtifacts","state":"succeeded"}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	status, err := c.WaitRunTerminal(context.Background(), "job#1", PollPolicy{InitialInterval: 5 * time.Millisecond})
	if err != nil {
		t.Fatalf("WaitRunTerminal() error = %v", err)
	}
	if status.GetState() != stateSucceeded {
		t.Fatalf("state = %s, want %s", status.GetState(), stateSucceeded)
	}
	if got := atomic.LoadInt32(&calls); got < 3 {
		t.Fatalf("calls = %d, want >= 3", got)
	}
}

func TestWaitRunTerminal_ContextCancelled(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"r-1","runId":"job#1","operation":"archiveArtifacts","state":"running"}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err = c.WaitRunTerminal(ctx, "job#1", PollPolicy{InitialInterval: 10 * time.Millisecond})
	if err == nil {
		t.Fatalf("WaitRunTerminal() error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("error = %v, want context deadline exceeded", err)
	}
}

func TestWaitRunTerminal_MaxAttempts(t *testing.T) {
	t.Parallel()

	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"r-1","runId":"job#1","operation":"archiveArtifacts","state":"running"}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.WaitRunTerminal(context.Background(), "job#1", PollPolicy{
		InitialInterval: time.Millisecond,
		MaxAttempts:     3,
	})
	if err == nil {
		t.Fatalf("WaitRunTerminal() error = nil, want max attempts error")
	}
	if !strings.Contains(err.Error(), "max attempts") {
		t.Fatalf("error = %v, want max attempts exhausted", err)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}

func TestWaitRunTerminal_MaxDuration(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"r-1","runId":"job#1","operation":"archiveArtifacts","state":"running"}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = c.WaitRunTerminal(context.Background(), "job#1", PollPolicy{
		InitialInterval: 5 * time.Millisecond,
		MaxDuration:     30 * time.Millisecond,
	})
	if err == nil {
		t.Fatalf("WaitRunTerminal() error = nil, want timeout error")
	}
	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("error = %v, want context deadline exceeded", err)
	}
}

func assertHTTPError(t *testing.T, err error, wantStatus int, wantCode string) {
	t.Helper()

	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error type = %T, want *HTTPError", err)
	}
	if httpErr.StatusCode != wantStatus {
		t.Fatalf("statusCode = %d, want %d", httpErr.StatusCode, wantStatus)
	}
	if httpErr.ProtoError == nil || httpErr.ProtoError.GetCode() != wantCode {
		t.Fatalf("proto error code = %v, want %s", httpErr.ProtoError, wantCode)
	}
}

func mustReadAll(t *testing.T, body io.Reader) []byte {
	t.Helper()
	out, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	return out
}
