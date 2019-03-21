package throttled

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// Transport will limit the requests to a base round tripper
type Transport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

// NewTransport thottled http transport
func NewTransport(base http.RoundTripper, limiter *rate.Limiter) *Transport {
	return &Transport{base, limiter}
}

// RoundTrip implementation with rate limiting
func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	res := t.limiter.Reserve()

	select {
	case <-time.After(res.Delay()):
		return t.base.RoundTrip(r)
	case <-r.Context().Done():
		res.Cancel()
		return nil, r.Context().Err()
	}
}
