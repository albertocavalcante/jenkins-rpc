package rpcclient

import "net/http"

// DebugHook provides callbacks for request/response inspection.
type DebugHook struct {
	// OnRequest is called before the HTTP request is sent.
	// req is the outgoing request; body is the request body (nil for GET).
	OnRequest func(req *http.Request, body []byte)

	// OnResponse is called after the HTTP response is received.
	// resp may be nil if the request failed at the transport level.
	OnResponse func(resp *http.Response, body []byte, err error)
}
