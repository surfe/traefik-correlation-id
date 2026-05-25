package traefik_correlation_id

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestCreateConfig(t *testing.T) {
	cfg := CreateConfig()
	if cfg.HeaderName != "correlation-id" {
		t.Errorf("default HeaderName = %q, want %q", cfg.HeaderName, "correlation-id")
	}
}

func TestNew_defaultsEmptyHeaderName(t *testing.T) {
	cfg := &Config{HeaderName: ""}
	h, err := New(context.Background(), http.NotFoundHandler(), cfg, "test")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	mid, ok := h.(*correlationID)
	if !ok {
		t.Fatal("expected *correlationID")
	}
	if mid.headerName != "correlation-id" {
		t.Errorf("headerName = %q, want %q", mid.headerName, "correlation-id")
	}
}

func TestServeHTTP_setsHeaderWhenAbsent(t *testing.T) {
	var capturedHeader string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("correlation-id") //nolint:canonicalheader
	})

	cfg := CreateConfig()
	h, _ := New(context.Background(), next, cfg, "test")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if capturedHeader == "" {
		t.Fatal("expected correlation ID header to be set, got empty string")
	}
	if !uuidRegex.MatchString(capturedHeader) {
		t.Errorf("generated ID %q is not a valid UUID v4", capturedHeader)
	}
}

func TestServeHTTP_preservesExistingHeader(t *testing.T) {
	const existing = "my-correlation-id"
	var capturedHeader string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("correlation-id") //nolint:canonicalheader
	})

	cfg := CreateConfig()
	h, _ := New(context.Background(), next, cfg, "test")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("correlation-id", existing) //nolint:canonicalheader
	h.ServeHTTP(httptest.NewRecorder(), req)

	if capturedHeader != existing {
		t.Errorf("header = %q, want %q", capturedHeader, existing)
	}
}

func TestServeHTTP_customHeaderName(t *testing.T) {
	var capturedHeader string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Request-Id")
	})

	cfg := &Config{HeaderName: "X-Request-Id"}
	h, _ := New(context.Background(), next, cfg, "test")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !uuidRegex.MatchString(capturedHeader) {
		t.Errorf("generated ID %q is not a valid UUID v4", capturedHeader)
	}
}

func TestServeHTTP_uniqueIDs(t *testing.T) {
	seen := make(map[string]bool)
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("correlation-id") //nolint:canonicalheader
		if seen[id] {
			t.Errorf("duplicate correlation ID generated: %q", id)
		}
		seen[id] = true
	})

	cfg := CreateConfig()
	h, _ := New(context.Background(), next, cfg, "test")

	for i := 0; i < 100; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
}

func TestNewUUID_format(t *testing.T) {
	for i := 0; i < 20; i++ {
		id := newUUID()
		if !uuidRegex.MatchString(id) {
			t.Errorf("newUUID() = %q, does not match UUID v4 pattern", id)
		}
	}
}
