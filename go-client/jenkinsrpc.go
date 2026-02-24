// Package jenkinsrpc provides a typed Go client for the Jenkins Step RPC plugin.
//
// The implementation lives in internal/rpcclient. This file re-exports public
// symbols so that external modules can import jenkinsrpc directly.
package jenkinsrpc

import (
	"net/http"

	"github.com/albertocavalcante/jenkins-rpc/go-client/internal/rpcclient"

	steprpcv1 "github.com/albertocavalcante/jenkins-rpc/contracts/gen/go/proto/steprpc/v1"
)

// Client is the HTTP client for the Jenkins Step RPC plugin API.
type Client = rpcclient.Client

// PollPolicy controls polling behavior for WaitRunTerminal.
type PollPolicy = rpcclient.PollPolicy

// RetryPolicy controls automatic retry behavior for transient failures.
type RetryPolicy = rpcclient.RetryPolicy

// DebugHook provides callbacks for request/response inspection.
type DebugHook = rpcclient.DebugHook

// HTTPError represents a non-2xx HTTP response with optional structured error details.
type HTTPError = rpcclient.HTTPError

// ErrorCategory classifies HTTP errors into broad operational categories.
type ErrorCategory = rpcclient.ErrorCategory

const (
	CategoryUnknown     = rpcclient.CategoryUnknown
	CategoryNetwork     = rpcclient.CategoryNetwork
	CategoryAuth        = rpcclient.CategoryAuth
	CategoryNotFound    = rpcclient.CategoryNotFound
	CategoryBadRequest  = rpcclient.CategoryBadRequest
	CategoryRateLimited = rpcclient.CategoryRateLimited
	CategoryServerError = rpcclient.CategoryServerError
)

// New creates a new client for the Jenkins Step RPC plugin.
func New(baseURL, token string, httpClient *http.Client) (*Client, error) {
	return rpcclient.New(baseURL, token, httpClient)
}

// DirectOperations returns catalog operations executable in the direct controller lane.
func DirectOperations(catalog *steprpcv1.CatalogResponse) []string {
	return rpcclient.DirectOperations(catalog)
}

// CPSBridgeOperations returns catalog operations requiring CPS bridge execution.
func CPSBridgeOperations(catalog *steprpcv1.CatalogResponse) []string {
	return rpcclient.CPSBridgeOperations(catalog)
}

// CategoryOf extracts the ErrorCategory from an error chain.
func CategoryOf(err error) ErrorCategory {
	return rpcclient.CategoryOf(err)
}
