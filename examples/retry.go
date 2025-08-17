package main

import (
	"context"
	"fmt"
	"math"
	"time"
)

type Request struct{}
type Response struct{}

var ErrMaxAttemptsExceeded = fmt.Errorf("max attempts reached")

type Client interface {
	DoSomething(ctx context.Context, req Request) (*Response, error)
}
type retryClient struct {
	Client                                Client
	MaxAttempts                           int           // e.g. 3
	Timeout, InitialInterval, MaxInterval time.Duration // e.g. 5s, 100ms, 1s
	Multiplier                            float64       // e.g. 2.0
}

func (r *retryClient) DoSomething(ctx context.Context, req Request) (*Response, error) {
	for i := 0; i < r.MaxAttempts; i++ {
		// Create a new context with timeout for each attempt...
		tryCtx, cancel := context.WithTimeout(ctx, r.Timeout)

		// Do underlying call...
		resp, err := r.Client.DoSomething(tryCtx, req)

		cancel()
		if err == nil {
			return resp, nil // Return response on success...
		}

		// Sleep before next attempt...
		if i < r.MaxAttempts-1 {
			// Wait for the backoff duration before the next attempt...
			<-time.After(r.backoffDelay(i))
		}
	}

	return nil, ErrMaxAttemptsExceeded
}

func (r *retryClient) backoffDelay(attempt int) time.Duration {
	ms := float64(r.InitialInterval) * math.Pow(r.Multiplier, float64(attempt))
	if r.MaxInterval > 0 && time.Duration(ms) > r.MaxInterval {
		return r.MaxInterval
	}
	return time.Duration(ms)
}
