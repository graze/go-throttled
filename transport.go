package throttled

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// Transport will limit the requests to a base round tripper
//
// It will apply the rate limiting supplied by `rate.Limiter` to all requests
// It observes the context.Done() channel to break out early if required
type Transport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

// Client returns a default http client with rate limiting
//
// Generates a default http.Client using this rate limiting transport
//
// Example:
//
//     throttled.Client(rate.NewLimiter(rate.Limit(4), 40))
func Client(limiter *rate.Limiter) *http.Client {
	return &http.Client{Transport: NewTransport(http.DefaultTransport, limiter)}
}

// NewTransport thottled http transport
//
// Generates a new Transport with rate limiting
//
// Example:
//
//     throttled.NewTransport(http.DefaultTransport, rate.NewLimiter(rate.Limit(4), 40))
func NewTransport(base http.RoundTripper, limiter *rate.Limiter) *Transport {
	return &Transport{base, limiter}
}

// WrapClient wraps an existing clients transport with the rate limiting transport
//
// Wraps an existing client with a new transport (useful for injected into third part clients)
//
// Example:
//
//     client.Client = throttled.WrapClient(client.Client, rate.NewLimiter(rate.Limit(4), 40))
func WrapClient(client *http.Client, limiter *rate.Limiter) *http.Client {
	if client == nil {
		client = &http.Client{Transport: http.DefaultTransport}
	}
	ts := client.Transport
	if ts == nil {
		ts = http.DefaultTransport
	}
	client.Transport = NewTransport(ts, limiter)
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
