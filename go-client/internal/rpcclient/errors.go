package rpcclient

import (
	"errors"
	"fmt"
	"net/http"

	steprpcv1 "github.com/albertocavalcante/jenkins-rpc/contracts/gen/go/proto/steprpc/v1"
)

// ErrorCategory classifies HTTP errors into broad operational categories.
type ErrorCategory int

const (
	CategoryUnknown     ErrorCategory = iota
	CategoryNetwork                   // connection/transport failures
	CategoryAuth                      // 401, 403
	CategoryNotFound                  // 404
	CategoryBadRequest                // 400
	CategoryRateLimited               // 429
	CategoryServerError               // 500-599
)

func (c ErrorCategory) String() string {
	switch c {
	case CategoryUnknown:
		return "Unknown"
	case CategoryNetwork:
		return "Network"
	case CategoryAuth:
		return "Auth"
	case CategoryNotFound:
		return "NotFound"
	case CategoryBadRequest:
		return "BadRequest"
	case CategoryRateLimited:
		return "RateLimited"
	case CategoryServerError:
		return "ServerError"
	default:
		return "Unknown"
	}
}

// HTTPError returns status code plus structured error details when available.
type HTTPError struct {
	StatusCode int
	ProtoError *steprpcv1.Error
}

func (e *HTTPError) Error() string {
	if e.ProtoError != nil && e.ProtoError.GetMessage() != "" {
		return fmt.Sprintf("request failed: status=%d code=%s message=%s", e.StatusCode, e.ProtoError.GetCode(), e.ProtoError.GetMessage())
	}
	return fmt.Sprintf("request failed: status=%d", e.StatusCode)
}

// Category returns the error category for this HTTP error.
func (e *HTTPError) Category() ErrorCategory {
	switch {
	case e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden:
		return CategoryAuth
	case e.StatusCode == http.StatusNotFound:
		return CategoryNotFound
	case e.StatusCode == http.StatusBadRequest:
		return CategoryBadRequest
	case e.StatusCode == http.StatusTooManyRequests:
		return CategoryRateLimited
	case e.StatusCode >= http.StatusInternalServerError:
		return CategoryServerError
	default:
		return CategoryUnknown
	}
}

// CategoryOf extracts the ErrorCategory from an error chain.
// Returns CategoryNetwork for non-HTTPError errors (transport failures)
// and CategoryUnknown if the error is nil.
func CategoryOf(err error) ErrorCategory {
	if err == nil {
		return CategoryUnknown
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.Category()
	}
	return CategoryNetwork
}
