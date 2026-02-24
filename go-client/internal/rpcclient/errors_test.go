package rpcclient

import (
	"fmt"
	"net/http"
	"testing"

	steprpcv1 "github.com/albertocavalcante/jenkins-rpc/contracts/gen/go/proto/steprpc/v1"
)

func TestHTTPError_Category(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		want       ErrorCategory
	}{
		{"400 → BadRequest", http.StatusBadRequest, CategoryBadRequest},
		{"401 → Auth", http.StatusUnauthorized, CategoryAuth},
		{"403 → Auth", http.StatusForbidden, CategoryAuth},
		{"404 → NotFound", http.StatusNotFound, CategoryNotFound},
		{"429 → RateLimited", http.StatusTooManyRequests, CategoryRateLimited},
		{"500 → ServerError", http.StatusInternalServerError, CategoryServerError},
		{"502 → ServerError", http.StatusBadGateway, CategoryServerError},
		{"503 → ServerError", http.StatusServiceUnavailable, CategoryServerError},
		{"409 → Unknown", http.StatusConflict, CategoryUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := &HTTPError{StatusCode: tt.statusCode}
			if got := err.Category(); got != tt.want {
				t.Fatalf("Category() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCategoryOf(t *testing.T) {
	t.Parallel()

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		if got := CategoryOf(nil); got != CategoryUnknown {
			t.Fatalf("CategoryOf(nil) = %v, want Unknown", got)
		}
	})

	t.Run("HTTPError", func(t *testing.T) {
		t.Parallel()
		err := fmt.Errorf("send: %w", &HTTPError{StatusCode: http.StatusNotFound})
		if got := CategoryOf(err); got != CategoryNotFound {
			t.Fatalf("CategoryOf(wrapped 404) = %v, want NotFound", got)
		}
	})

	t.Run("non-HTTP error → Network", func(t *testing.T) {
		t.Parallel()
		err := fmt.Errorf("dial tcp: connection refused")
		if got := CategoryOf(err); got != CategoryNetwork {
			t.Fatalf("CategoryOf(dial error) = %v, want Network", got)
		}
	})
}

func TestHTTPError_Error(t *testing.T) {
	t.Parallel()

	t.Run("with proto error", func(t *testing.T) {
		t.Parallel()
		err := &HTTPError{
			StatusCode: http.StatusBadRequest,
			ProtoError: &steprpcv1.Error{Code: "bad_op", Message: "denied"},
		}
		want := "request failed: status=400 code=bad_op message=denied"
		if got := err.Error(); got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("without proto error", func(t *testing.T) {
		t.Parallel()
		err := &HTTPError{StatusCode: http.StatusInternalServerError}
		want := "request failed: status=500"
		if got := err.Error(); got != want {
			t.Fatalf("Error() = %q, want %q", got, want)
		}
	})
}

func TestErrorCategory_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		cat  ErrorCategory
		want string
	}{
		{CategoryUnknown, "Unknown"},
		{CategoryNetwork, "Network"},
		{CategoryAuth, "Auth"},
		{CategoryNotFound, "NotFound"},
		{CategoryBadRequest, "BadRequest"},
		{CategoryRateLimited, "RateLimited"},
		{CategoryServerError, "ServerError"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.cat.String(); got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
