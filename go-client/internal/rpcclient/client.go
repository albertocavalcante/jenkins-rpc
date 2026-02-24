package rpcclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strings"
	"time"

	steprpcv1 "github.com/albertocavalcante/jenkins-rpc/contracts/gen/go/proto/steprpc/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const defaultPollInterval = 2 * time.Second

// PollPolicy controls polling behavior for WaitRunTerminal.
type PollPolicy struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxAttempts     int
	MaxDuration     time.Duration
}

// Client is a minimal HTTP client scaffold for the plugin RPC API.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	token       string
	retryPolicy *RetryPolicy
	debugHook   *DebugHook
}

// New creates a new client scaffold.
func New(baseURL, token string, httpClient *http.Client) (*Client, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("baseURL is required")
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
		token:      token,
	}, nil
}

// WithRetryPolicy returns a copy of the client with the given retry policy.
func (c *Client) WithRetryPolicy(p *RetryPolicy) *Client {
	cp := *c
	cp.retryPolicy = p
	return &cp
}

// WithDebugHook returns a copy of the client with the given debug hook.
func (c *Client) WithDebugHook(h *DebugHook) *Client {
	cp := *c
	cp.debugHook = h
	return &cp
}

// Invoke sends an invoke request to the plugin.
func (c *Client) Invoke(ctx context.Context, req *steprpcv1.InvokeRequest) (*steprpcv1.InvokeResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("invoke request is required")
	}

	payload, err := protojson.MarshalOptions{
		UseProtoNames: false,
	}.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal invoke request: %w", err)
	}

	body, err := c.doRequestWithRetry(ctx, func() (*http.Request, error) {
		httpReq, reqErr := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.baseURL+"/step-rpc/v1/invoke",
			bytes.NewReader(payload),
		)
		if reqErr != nil {
			return nil, reqErr
		}
		httpReq.Header.Set("Content-Type", "application/json")
		if c.token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+c.token)
		}
		return httpReq, nil
	})
	if err != nil {
		return nil, fmt.Errorf("send invoke request: %w", err)
	}

	out := &steprpcv1.InvokeResponse{}
	if err := protojson.Unmarshal(body, out); err != nil {
		return nil, fmt.Errorf("decode invoke response: %w", err)
	}
	return out, nil
}

// GetRunStatus fetches status for a run ID.
func (c *Client) GetRunStatus(ctx context.Context, runID string) (*steprpcv1.RunStatusResponse, error) {
	if strings.TrimSpace(runID) == "" {
		return nil, fmt.Errorf("runID is required")
	}

	out := &steprpcv1.RunStatusResponse{}
	if err := c.getProto(ctx, "/step-rpc/v1/runs/"+url.PathEscape(runID), out, "status"); err != nil {
		return nil, err
	}
	return out, nil
}

// GetCatalog fetches operation discovery metadata from the plugin.
func (c *Client) GetCatalog(ctx context.Context) (*steprpcv1.CatalogResponse, error) {
	out := &steprpcv1.CatalogResponse{}
	if err := c.getProto(ctx, "/step-rpc/v1/catalog", out, "catalog"); err != nil {
		return nil, err
	}
	return out, nil
}

// WaitRunTerminal polls run status until a terminal state is reached or context is canceled.
func (c *Client) WaitRunTerminal(ctx context.Context, runID string, policy PollPolicy) (*steprpcv1.RunStatusResponse, error) {
	if strings.TrimSpace(runID) == "" {
		return nil, fmt.Errorf("runID is required")
	}
	if policy.InitialInterval <= 0 {
		policy.InitialInterval = defaultPollInterval
	}
	if policy.MaxInterval <= 0 {
		policy.MaxInterval = policy.InitialInterval
	}

	var cancel context.CancelFunc
	if policy.MaxDuration > 0 {
		ctx, cancel = context.WithTimeout(ctx, policy.MaxDuration)
		defer cancel()
	}

	interval := policy.InitialInterval
	for attempt := 0; ; attempt++ {
		status, err := c.GetRunStatus(ctx, runID)
		if err != nil {
			return nil, err
		}
		if isTerminalState(status.GetState()) {
			return status, nil
		}

		if policy.MaxAttempts > 0 && attempt+1 >= policy.MaxAttempts {
			return nil, fmt.Errorf("wait run terminal: max attempts (%d) exhausted", policy.MaxAttempts)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait run terminal: %w", ctx.Err())
		case <-time.After(pollJitter(interval)):
		}

		// Exponential backoff for next iteration.
		interval *= 2
		if interval > policy.MaxInterval {
			interval = policy.MaxInterval
		}
	}
}

func pollJitter(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	// Apply +-25% jitter.
	quarter := int64(d) / 4 //nolint:mnd // 25% jitter range
	if quarter <= 0 {
		return d
	}
	return d - time.Duration(quarter) + time.Duration(rand.Int64N(2*quarter+1)) //nolint:gosec // jitter does not need cryptographic randomness
}

// DirectOperations returns catalog operations executable in the direct controller lane.
func DirectOperations(catalog *steprpcv1.CatalogResponse) []string {
	return operationsByExecutionMode(catalog, steprpcv1.OperationExecutionMode_OPERATION_EXECUTION_MODE_DIRECT)
}

// CPSBridgeOperations returns catalog operations requiring CPS bridge execution.
func CPSBridgeOperations(catalog *steprpcv1.CatalogResponse) []string {
	return operationsByExecutionMode(catalog, steprpcv1.OperationExecutionMode_OPERATION_EXECUTION_MODE_CPS_BRIDGE_REQUIRED)
}

func operationsByExecutionMode(catalog *steprpcv1.CatalogResponse, mode steprpcv1.OperationExecutionMode) []string {
	if catalog == nil {
		return nil
	}
	var out []string
	for _, op := range catalog.GetOperations() {
		if op.GetExecutionMode() == mode {
			out = append(out, op.GetName())
		}
	}
	return out
}

func isTerminalState(state string) bool {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "succeeded", "failed", "cancelled":
		return true
	default:
		return false
	}
}

// GetBridgePending fetches the next pending CPS bridge request for a run.
func (c *Client) GetBridgePending(ctx context.Context, runExternalizableID string) (*steprpcv1.BridgePendingResponse, error) {
	if strings.TrimSpace(runExternalizableID) == "" {
		return nil, fmt.Errorf("runExternalizableID is required")
	}

	out := &steprpcv1.BridgePendingResponse{}
	endpoint := "/step-rpc/v1/bridge/pending?runExternalizableId=" + url.QueryEscape(runExternalizableID)
	if err := c.getProto(ctx, endpoint, out, "bridge pending"); err != nil {
		return nil, err
	}
	return out, nil
}

// CompleteBridgeRequest marks a CPS bridge request as terminal.
func (c *Client) CompleteBridgeRequest(ctx context.Context, req *steprpcv1.BridgeCompleteRequest) (*steprpcv1.BridgeCompleteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("bridge complete request is required")
	}
	if strings.TrimSpace(req.GetRunId()) == "" {
		return nil, fmt.Errorf("runID is required")
	}
	if strings.TrimSpace(req.GetState()) == "" {
		return nil, fmt.Errorf("state is required")
	}

	payload, err := protojson.MarshalOptions{
		UseProtoNames: false,
	}.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal bridge completion request: %w", err)
	}

	body, err := c.doRequestWithRetry(ctx, func() (*http.Request, error) {
		httpReq, reqErr := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			c.baseURL+"/step-rpc/v1/bridge/complete",
			bytes.NewReader(payload),
		)
		if reqErr != nil {
			return nil, reqErr
		}
		httpReq.Header.Set("Content-Type", "application/json")
		if c.token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+c.token)
		}
		return httpReq, nil
	})
	if err != nil {
		return nil, fmt.Errorf("send bridge completion request: %w", err)
	}

	out := &steprpcv1.BridgeCompleteResponse{}
	if err := protojson.Unmarshal(body, out); err != nil {
		return nil, fmt.Errorf("decode bridge completion response: %w", err)
	}
	return out, nil
}

func (c *Client) doRequest(httpReq *http.Request) (statusCode int, body []byte, err error) {
	if c.debugHook != nil && c.debugHook.OnRequest != nil {
		var reqBody []byte
		if httpReq.Body != nil && httpReq.GetBody != nil {
			if r, cloneErr := httpReq.GetBody(); cloneErr == nil {
				reqBody, _ = io.ReadAll(r)
			}
		}
		c.debugHook.OnRequest(httpReq, reqBody)
	}

	httpResp, doErr := c.httpClient.Do(httpReq)
	if doErr != nil {
		if c.debugHook != nil && c.debugHook.OnResponse != nil {
			c.debugHook.OnResponse(nil, nil, doErr)
		}
		return 0, nil, doErr
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	body, readErr := io.ReadAll(httpResp.Body)
	if readErr != nil {
		readErr = fmt.Errorf("read response body: %w", readErr)
		if c.debugHook != nil && c.debugHook.OnResponse != nil {
			c.debugHook.OnResponse(httpResp, nil, readErr)
		}
		return httpResp.StatusCode, nil, readErr
	}

	if httpResp.StatusCode < http.StatusOK || httpResp.StatusCode >= http.StatusMultipleChoices {
		httpErr := &HTTPError{StatusCode: httpResp.StatusCode}
		errResp := &steprpcv1.ErrorResponse{}
		if unmarshalErr := protojson.Unmarshal(body, errResp); unmarshalErr == nil && errResp.GetError() != nil {
			httpErr.ProtoError = errResp.GetError()
		}
		if c.debugHook != nil && c.debugHook.OnResponse != nil {
			c.debugHook.OnResponse(httpResp, body, httpErr)
		}
		return httpResp.StatusCode, nil, httpErr
	}

	if c.debugHook != nil && c.debugHook.OnResponse != nil {
		c.debugHook.OnResponse(httpResp, body, nil)
	}
	return httpResp.StatusCode, body, nil
}

func (c *Client) doRequestWithRetry(ctx context.Context, buildReq func() (*http.Request, error)) ([]byte, error) {
	return doWithRetry(ctx, c.retryPolicy, func() (int, []byte, error) {
		httpReq, err := buildReq()
		if err != nil {
			return 0, nil, err
		}
		return c.doRequest(httpReq)
	})
}

func (c *Client) getProto(ctx context.Context, endpoint string, out proto.Message, name string) error {
	body, err := c.doRequestWithRetry(ctx, func() (*http.Request, error) {
		httpReq, reqErr := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			c.baseURL+endpoint,
			nil,
		)
		if reqErr != nil {
			return nil, fmt.Errorf("build %s request: %w", name, reqErr)
		}
		if c.token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+c.token)
		}
		return httpReq, nil
	})
	if err != nil {
		return fmt.Errorf("send %s request: %w", name, err)
	}
	if unmarshalErr := protojson.Unmarshal(body, out); unmarshalErr != nil {
		return fmt.Errorf("decode %s response: %w", name, unmarshalErr)
	}

	return nil
}
