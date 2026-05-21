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
	if cfg.HeaderName != "X-Correlation-Id" {
		t.Errorf("default HeaderName = %q, want %q", cfg.HeaderName, "X-Correlation-Id")
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
	if mid.headerName != "X-Correlation-Id" {
		t.Errorf("headerName = %q, want %q", mid.headerName, "X-Correlation-Id")
	}
}

func TestServeHTTP_setsHeaderWhenAbsent(t *testing.T) {
	var capturedHeader string
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Correlation-Id")
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
		capturedHeader = r.Header.Get("X-Correlation-Id")
	})

	cfg := CreateConfig()
	h, _ := New(context.Background(), next, cfg, "test")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-Id", existing)
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
		id := r.Header.Get("X-Correlation-Id")
		if seen[id] {
			t.Errorf("duplicate correlation ID generated: %q", id)
		}
		seen[id] = true
	})

	cfg := CreateConfig()
	h, _ := New(context.Background(), next, cfg, "test")

	for range 100 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(httptest.NewRecorder(), req)
	}
}

func TestNewUUID_format(t *testing.T) {
	for range 20 {
		id := newUUID()
		if !uuidRegex.MatchString(id) {
			t.Errorf("newUUID() = %q, does not match UUID v4 pattern", id)
		}
	}
}
