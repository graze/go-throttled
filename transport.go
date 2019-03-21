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

// WrapClient wraps an existing clients transport with the rate limiting transport
func WrapClient(client *http.Client, limiter *rate.Limiter) *http.Client {
	client.Transport = NewTransport(client.Transport, limiter)
	return client
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
