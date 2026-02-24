package rpcclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	steprpcv1 "github.com/albertocavalcante/jenkins-rpc/contracts/gen/go/proto/steprpc/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func retryClient(t *testing.T, ts *httptest.Server, maxAttempts int) *Client {
	t.Helper()
	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return c.WithRetryPolicy(&RetryPolicy{
		MaxAttempts:    maxAttempts,
		InitialBackoff: time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
	})
}

func TestRetry_503ThenSuccess(t *testing.T) {
	t.Parallel()

	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":{"code":"unavailable","message":"retry later"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"r-1","runId":"run-1","state":"queued"}`))
	}))
	defer ts.Close()

	c := retryClient(t, ts, 5)

	args, _ := structpb.NewStruct(map[string]any{"a": "b"})
	resp, err := c.Invoke(context.Background(), &steprpcv1.InvokeRequest{
		RequestId: "r-1",
		Operation: "test",
		Args:      args,
	})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}
	if resp.GetRunId() != "run-1" {
		t.Fatalf("runId = %s, want run-1", resp.GetRunId())
	}
	if got := atomic.LoadInt32(&calls); got < 3 {
		t.Fatalf("calls = %d, want >= 3", got)
	}
}

func TestRetry_NoRetryOn400(t *testing.T) {
	t.Parallel()

	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"bad_request","message":"nope"}}`))
	}))
	defer ts.Close()

	c := retryClient(t, ts, 3)

	_, err := c.Invoke(context.Background(), &steprpcv1.InvokeRequest{
		RequestId: "r-1",
		Operation: "test",
	})
	if err == nil {
		t.Fatalf("Invoke() error = nil, want non-nil")
	}
	assertHTTPError(t, err, http.StatusBadRequest, "bad_request")
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("calls = %d, want 1 (no retry on 400)", got)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":{"code":"unavailable","message":"retry"}}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	c = c.WithRetryPolicy(&RetryPolicy{
		MaxAttempts:    100,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = c.Invoke(ctx, &steprpcv1.InvokeRequest{
		RequestId: "r-1",
		Operation: "test",
	})
	if err == nil {
		t.Fatalf("Invoke() error = nil, want context error")
	}
}

func TestRetry_MaxAttemptsExhausted(t *testing.T) {
	t.Parallel()

	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":{"code":"unavailable","message":"still down"}}`))
	}))
	defer ts.Close()

	c := retryClient(t, ts, 3)

	_, err := c.Invoke(context.Background(), &steprpcv1.InvokeRequest{
		RequestId: "r-1",
		Operation: "test",
	})
	if err == nil {
		t.Fatalf("Invoke() error = nil, want non-nil")
	}
	assertHTTPError(t, err, http.StatusServiceUnavailable, "unavailable")
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("calls = %d, want 3", got)
	}
}
