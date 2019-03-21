package throttled

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

type mockRoundTripper struct {
	count int
	rate  float64
	start *time.Time
}

// RoundTrip implementation
func (m *mockRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if m.start == nil {
		t := time.Now()
		m.start = &t
	}
	m.count++
	fmt.Println("count", m.count, "duration", time.Since(*m.start), "duration in seconds", float64(time.Since(*m.start))/float64(time.Second), "rate", float64(m.count)/(float64(time.Since(*m.start))/float64(time.Second)))
	m.rate = float64(m.count) / (float64(time.Since(*m.start)) / float64(time.Second))
	return &http.Response{}, nil
}

func TestTransport_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		limiter  *rate.Limiter
		wantRate float64
	}{
		{
			name:     "10/s",
			limiter:  rate.NewLimiter(rate.Limit(10), 1),
			wantRate: 10,
		},
		{
			name:     "100/s",
			limiter:  rate.NewLimiter(rate.Limit(100), 1),
			wantRate: 100,
		},
		{
			name:     "1000/s",
			limiter:  rate.NewLimiter(rate.Limit(1000), 1),
			wantRate: 1000,
		},
		{
			name:     "10/s 20 bucket",
			limiter:  rate.NewLimiter(rate.Limit(100), 20),
			wantRate: 160, // very variable!
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mrt := &mockRoundTripper{count: -1}
			transport := &Transport{
				base:    mrt,
				limiter: tt.limiter,
			}
			wg := &sync.WaitGroup{}
			for i := 0; i < 50; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					transport.RoundTrip(&http.Request{})
				}()
			}
			wg.Wait()

			// compare the percentage difference
			assert.InDelta(t, math.Abs(tt.wantRate/mrt.rate), 1, 0.1)
		})
	}
}

func TestTransport_ContextCancel(t *testing.T) {
	mrt := &mockRoundTripper{count: -1}
	transport := &Transport{
		base:    mrt,
		limiter: rate.NewLimiter(rate.Limit(1), 0), // never execute limiter
	}
	ctx, cancel := context.WithCancel(context.Background())
	r, err := http.NewRequest("GET", "/path", nil)
	r = r.WithContext(ctx)
	require.Nil(t, err)

	hit := false
	c := make(chan bool)
	go func() {
		_, err := transport.RoundTrip(r)
		hit = true
		require.NotNil(t, err)
		assert.Equal(t, "context canceled", err.Error())
		c <- true
	}()

	// cancel the context
	cancel()

	<-c

	assert.True(t, hit, "the context was cancelled")
}

func TestTransport_ContextTimeout(t *testing.T) {
	mrt := &mockRoundTripper{count: -1}
	transport := &Transport{
		base:    mrt,
		limiter: rate.NewLimiter(rate.Limit(1), 0), // never execute limiter
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	r, err := http.NewRequest("GET", "/path", nil)
	r = r.WithContext(ctx)
	require.Nil(t, err)

	hit := false
	c := make(chan bool)
	go func() {
		_, err := transport.RoundTrip(r)
		hit = true
		require.NotNil(t, err)
		assert.Equal(t, "context deadline exceeded", err.Error())
		c <- true
	}()

	<-c

	// cancel the context after the timeout
	cancel()

	assert.True(t, hit, "the context was cancelled")
}

func TestWrapClient(t *testing.T) {
	tests := []struct {
		name    string
		client  *http.Client
		limiter *rate.Limiter
		want    *http.Client
	}{
		{
			name:    "has transport",
			client:  &http.Client{Transport: http.DefaultTransport},
			limiter: rate.NewLimiter(rate.Limit(10), 1),
			want:    &http.Client{Transport: &Transport{base: http.DefaultTransport, limiter: rate.NewLimiter(rate.Limit(10), 1)}},
		},
		{
			name:    "no transport",
			client:  &http.Client{},
			limiter: rate.NewLimiter(rate.Limit(10), 1),
			want:    &http.Client{Transport: &Transport{base: http.DefaultTransport, limiter: rate.NewLimiter(rate.Limit(10), 1)}},
		},
		{
			name:    "nil client",
			client:  nil,
			limiter: rate.NewLimiter(rate.Limit(10), 1),
			want:    &http.Client{Transport: &Transport{base: http.DefaultTransport, limiter: rate.NewLimiter(rate.Limit(10), 1)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapClient(tt.client, tt.limiter)
			assert.Equal(t, tt.want, got)
		})
	}
}
