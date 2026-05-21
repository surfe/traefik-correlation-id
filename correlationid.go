package traefik_correlation_id

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

// Config holds the plugin configuration.
type Config struct {
	// HeaderName is the header used to carry the correlation ID (default: X-Correlation-Id).
	HeaderName string `json:"headerName,omitempty"`
}

// CreateConfig returns the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		HeaderName: "X-Correlation-Id",
	}
}

type correlationID struct {
	next       http.Handler
	headerName string
	name       string
}

// New creates a new correlation-Id middleware instance.
func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	if cfg.HeaderName == "" {
		cfg.HeaderName = "X-Correlation-Id"
	}
	return &correlationID{
		next:       next,
		headerName: cfg.HeaderName,
		name:       name,
	}, nil
}

func (c *correlationID) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Header.Get(c.headerName) == "" {
		req.Header.Set(c.headerName, newUUID())
	}

	c.next.ServeHTTP(rw, req)
}

// newUUID returns a random UUID v4 string.
// Uses crypto/rand — no external dependency required.
func newUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:],
	)
}
