package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitLoggingTransport_RateLimitHit(t *testing.T) {
	// Create a test server that returns 429.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "1000")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(time.Now().Add(5*time.Second).Unix())))
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	metrics := &rateLimitMetrics{minRemaining: -1}
	transport := newRateLimitLoggingTransport(http.DefaultTransport, metrics)
	client := &http.Client{Transport: transport}

	req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	assert.Equal(t, 1, metrics.totalRequests)
	assert.Equal(t, 1, metrics.rateLimitHits)
	assert.True(t, metrics.totalWaitDuration > 0)
	assert.Equal(t, 0, metrics.minRemaining)
}

func TestRateLimitLoggingTransport_ProximityWarning_5Percent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "1000")
		w.Header().Set("X-RateLimit-Remaining", "50")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	metrics := &rateLimitMetrics{minRemaining: -1}
	transport := newRateLimitLoggingTransport(http.DefaultTransport, metrics)
	client := &http.Client{Transport: transport}

	req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	assert.Equal(t, 1, metrics.totalRequests)
	assert.Equal(t, 0, metrics.rateLimitHits)
	assert.Equal(t, 50, metrics.minRemaining)
}

func TestRateLimitLoggingTransport_ProximityWarning_10Percent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "1000")
		w.Header().Set("X-RateLimit-Remaining", "100")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	metrics := &rateLimitMetrics{minRemaining: -1}
	transport := newRateLimitLoggingTransport(http.DefaultTransport, metrics)
	client := &http.Client{Transport: transport}

	req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 100, metrics.minRemaining)
}

func TestRateLimitLoggingTransport_NoWarning_Healthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Limit", "1000")
		w.Header().Set("X-RateLimit-Remaining", "500")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	metrics := &rateLimitMetrics{minRemaining: -1}
	transport := newRateLimitLoggingTransport(http.DefaultTransport, metrics)
	client := &http.Client{Transport: transport}

	req, err := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 500, metrics.minRemaining)
	assert.Equal(t, 500, metrics.maxRemaining)
}

func TestRateLimitMetrics_RecordMultipleRequests(t *testing.T) {
	metrics := &rateLimitMetrics{minRemaining: -1}

	// Record several requests with different remaining values.
	metrics.recordRequest(1000, false, 0)
	metrics.recordRequest(950, false, 0)
	metrics.recordRequest(900, false, 0)
	metrics.recordRequest(50, false, 0)           // Lowest.
	metrics.recordRequest(0, true, 5*time.Second) // Rate limited.

	assert.Equal(t, 5, metrics.totalRequests)
	assert.Equal(t, 1, metrics.rateLimitHits)
	assert.Equal(t, 5*time.Second, metrics.totalWaitDuration)
	assert.Equal(t, 0, metrics.minRemaining)
	assert.Equal(t, 1000, metrics.maxRemaining)
}

func TestRateLimitMetrics_EfficiencyScore(t *testing.T) {
	tests := []struct {
		name               string
		totalRequests      int
		rateLimitHits      int
		expectedEfficiency float64
	}{
		{
			name:               "Perfect efficiency (no rate limits)",
			totalRequests:      100,
			rateLimitHits:      0,
			expectedEfficiency: 100.0,
		},
		{
			name:               "One rate limit hit",
			totalRequests:      100,
			rateLimitHits:      1,
			expectedEfficiency: 99.0,
		},
		{
			name:               "Multiple rate limit hits",
			totalRequests:      100,
			rateLimitHits:      5,
			expectedEfficiency: 95.0,
		},
		{
			name:               "High rate limit hits",
			totalRequests:      100,
			rateLimitHits:      25,
			expectedEfficiency: 75.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &rateLimitMetrics{
				totalRequests: tt.totalRequests,
				rateLimitHits: tt.rateLimitHits,
				minRemaining:  -1,
			}

			efficiencyScore := 100.0
			if metrics.totalRequests > 0 {
				efficiencyScore = 100.0 - (float64(metrics.rateLimitHits)/float64(metrics.totalRequests))*100.0
			}

			assert.InDelta(t, tt.expectedEfficiency, efficiencyScore, 0.1)
		})
	}
}

func TestRateLimitMetrics_Summary_EmptyMetrics(t *testing.T) {
	// Test that summary handles empty metrics gracefully.
	metrics := &rateLimitMetrics{minRemaining: -1}
	ctx := context.Background()

	// Should not panic with zero requests.
	metrics.logSummary(ctx)

	assert.Equal(t, 0, metrics.totalRequests)
}

func TestRateLimitMetrics_ConcurrentAccess(t *testing.T) {
	// Test that metrics are safe for concurrent access.
	metrics := &rateLimitMetrics{minRemaining: -1}

	// Simulate concurrent requests.
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(remaining int) {
			metrics.recordRequest(remaining, false, 0)
			done <- true
		}(i * 100)
	}

	// Wait for all goroutines.
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(t, 10, metrics.totalRequests)
	assert.Equal(t, 0, metrics.rateLimitHits)
}

func TestRateLimitLoggingTransport_Integration(t *testing.T) {
	// Integration test: multiple requests with varying rate limit statuses.
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		remaining := 1000 - (callCount * 100)

		w.Header().Set("X-RateLimit-Limit", "1000")
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if remaining <= 0 {
			w.Header().Set("X-RateLimit-Reset", strconv.Itoa(int(time.Now().Add(1*time.Second).Unix())))
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	metrics := &rateLimitMetrics{minRemaining: -1}
	transport := newRateLimitLoggingTransport(http.DefaultTransport, metrics)
	client := &http.Client{Transport: transport}

	// Make multiple requests.
	for i := 0; i < 12; i++ {
		req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
		resp, err := client.Do(req)
		require.NoError(t, err)
		_ = resp.Body.Close()
	}

	assert.Equal(t, 12, metrics.totalRequests)
	assert.Greater(t, metrics.rateLimitHits, 0) // Should have hit rate limit.
	assert.Equal(t, 0, metrics.minRemaining)    // Should have reached 0.
}
