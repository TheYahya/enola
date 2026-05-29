package enola

import (
	"net/http"
	"time"

	"golang.org/x/sync/semaphore"
)

const (
	defaultConcurrency = 20
	defaultTimeout     = 20 * time.Second
)

// Option configures an Enola instance.
type Option func(*Enola)

// WithHTTPClient sets the HTTP client used for probes.
func WithHTTPClient(client *http.Client) Option {
	return func(e *Enola) {
		e.client = client
	}
}

// WithConcurrency sets the maximum number of concurrent probes.
func WithConcurrency(n int64) Option {
	return func(e *Enola) {
		if n > 0 {
			e.sem = semaphore.NewWeighted(n)
		}
	}
}

// WithData replaces the embedded site database (mainly for tests).
func WithData(data map[string]Website) Option {
	return func(e *Enola) {
		e.Data = data
	}
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: defaultTimeout,
	}
}
