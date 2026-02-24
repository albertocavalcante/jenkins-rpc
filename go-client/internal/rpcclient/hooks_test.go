package rpcclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	steprpcv1 "github.com/albertocavalcante/jenkins-rpc/contracts/gen/go/proto/steprpc/v1"
)

func TestDebugHook_FiresOnRequestAndResponse(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"r-1","runId":"run-1","state":"queued"}`))
	}))
	defer ts.Close()

	var reqCalls, respCalls int32
	var capturedReqMethod string
	var capturedRespStatus int

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	c = c.WithDebugHook(&DebugHook{
		OnRequest: func(req *http.Request, _ []byte) {
			atomic.AddInt32(&reqCalls, 1)
			capturedReqMethod = req.Method
		},
		OnResponse: func(resp *http.Response, _ []byte, _ error) {
			atomic.AddInt32(&respCalls, 1)
			if resp != nil {
				capturedRespStatus = resp.StatusCode
			}
		},
	})

	_, err = c.Invoke(context.Background(), &steprpcv1.InvokeRequest{
		RequestId: "r-1",
		Operation: "test",
	})
	if err != nil {
		t.Fatalf("Invoke() error = %v", err)
	}

	if got := atomic.LoadInt32(&reqCalls); got != 1 {
		t.Fatalf("OnRequest called %d times, want 1", got)
	}
	if got := atomic.LoadInt32(&respCalls); got != 1 {
		t.Fatalf("OnResponse called %d times, want 1", got)
	}
	if capturedReqMethod != http.MethodPost {
		t.Fatalf("request method = %s, want POST", capturedReqMethod)
	}
	if capturedRespStatus != http.StatusOK {
		t.Fatalf("response status = %d, want 200", capturedRespStatus)
	}
}

func TestDebugHook_FiresOnError(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"code":"internal","message":"boom"}}`))
	}))
	defer ts.Close()

	var capturedErr error

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	c = c.WithDebugHook(&DebugHook{
		OnRequest:  func(_ *http.Request, _ []byte) {},
		OnResponse: func(_ *http.Response, _ []byte, err error) { capturedErr = err },
	})

	_, err = c.GetCatalog(context.Background())
	if err == nil {
		t.Fatalf("GetCatalog() error = nil, want error")
	}
	if capturedErr == nil {
		t.Fatalf("OnResponse error = nil, want non-nil")
	}
}

func TestDebugHook_NilCallbacksAreSkipped(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"operations":[]}`))
	}))
	defer ts.Close()

	c, err := New(ts.URL, "", ts.Client())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	// Hook with nil callbacks should not panic.
	c = c.WithDebugHook(&DebugHook{})

	_, err = c.GetCatalog(context.Background())
	if err != nil {
		t.Fatalf("GetCatalog() error = %v", err)
	}
}
