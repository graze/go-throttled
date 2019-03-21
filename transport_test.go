package throttled

import (
	"fmt"
	"math"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

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
					fmt.Println("calling round trip")
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
