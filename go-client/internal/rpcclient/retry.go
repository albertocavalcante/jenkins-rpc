package rpcclient

import (
	"context"
	"math"
	"math/rand/v2"
	"net/http"
	"time"
)

// RetryPolicy controls automatic retry behavior for transient failures.
// Zero value means no retries.
type RetryPolicy struct {
	MaxAttempts    int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	// Classifier overrides the default retryable check. Return true to retry.
	Classifier func(statusCode int, err error) bool
}

func (p *RetryPolicy) isRetryable(statusCode int, err error) bool {
	if p.Classifier != nil {
		return p.Classifier(statusCode, err)
	}
	return defaultRetryClassifier(statusCode, err)
}

func defaultRetryClassifier(statusCode int, _ error) bool {
	switch statusCode {
	case http.StatusTooManyRequests,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (p *RetryPolicy) backoff(attempt int) time.Duration {
	if p.InitialBackoff <= 0 {
		return 0
	}
	//nolint:mnd // exponent for exponential backoff
	dur := time.Duration(float64(p.InitialBackoff) * math.Pow(2, float64(attempt)))
	if p.MaxBackoff > 0 && dur > p.MaxBackoff {
		dur = p.MaxBackoff
	}
	// Full jitter: uniform random in [0, dur].
	if dur > 0 {
		dur = time.Duration(rand.Int64N(int64(dur) + 1)) //nolint:gosec // jitter does not need cryptographic randomness
	}
	return dur
}

// doWithRetry executes fn up to policy.MaxAttempts times, sleeping between retries.
// fn must return (statusCode, responseBody, error). A zero statusCode means the
// request never got a response (transport error).
func doWithRetry(ctx context.Context, policy *RetryPolicy, fn func() (int, []byte, error)) ([]byte, error) {
	if policy == nil || policy.MaxAttempts <= 1 {
		_, body, err := fn()
		return body, err
	}

	var lastErr error
	for attempt := range policy.MaxAttempts {
		statusCode, body, err := fn()
		if err == nil {
			return body, nil
		}
		lastErr = err

		if !policy.isRetryable(statusCode, err) {
			return nil, err
		}

		if attempt == policy.MaxAttempts-1 {
			break
		}

		wait := policy.backoff(attempt)
		if wait > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
		}
	}

	return nil, lastErr
}
