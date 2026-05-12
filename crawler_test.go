package crawler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	aeo "github.com/mizcausevic-dev/aeo-sdk-go"
)

// newTestServer returns an httptest server that serves a fixed AEO
// document at /.well-known/aeo.json and 404s everything else.
func newTestServer(t *testing.T, doc map[string]interface{}) *httptest.Server {
	t.Helper()
	body, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/aeo.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	})
	return httptest.NewServer(mux)
}

func TestCrawl_SeedOnly(t *testing.T) {
	srv := newTestServer(t, map[string]interface{}{
		"aeo_version": "0.1",
		"entity": map[string]interface{}{
			"id":            "https://example.com/#org",
			"type":          "Organization",
			"name":          "Solo Origin",
			"canonical_url": "https://example.com/",
		},
		"authority": map[string]interface{}{
			"primary_sources": []string{"https://example.com/"},
		},
		"claims": []map[string]interface{}{
			{"id": "tagline", "predicate": "description", "value": "test", "confidence": "high"},
		},
	})
	defer srv.Close()

	c := New(Config{MaxDepth: 0, MaxFetches: 10, Concurrency: 2, FetchTimeout: time.Second})
	results, err := c.Crawl(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Crawl: %v", err)
	}
	if got := len(results); got != 1 {
		t.Fatalf("len(results) = %d, want 1", got)
	}
	if !results[0].Success {
		t.Errorf("seed should succeed, got error: %s", results[0].Error)
	}
	if results[0].EntityName != "Solo Origin" {
		t.Errorf("EntityName = %q", results[0].EntityName)
	}
}

func TestCrawl_FollowsPrimarySources(t *testing.T) {
	// Set up two origins: the seed lists the second one as a primary source.
	var srv2 *httptest.Server
	srv2 = newTestServer(t, map[string]interface{}{
		"aeo_version": "0.1",
		"entity": map[string]interface{}{
			"id":            "https://child.example.com/#org",
			"type":          "Organization",
			"name":          "Child Origin",
			"canonical_url": "https://child.example.com/",
		},
		"authority": map[string]interface{}{
			"primary_sources": []string{"https://child.example.com/"},
		},
		"claims": []map[string]interface{}{
			{"id": "tagline", "predicate": "description", "value": "child", "confidence": "high"},
		},
	})
	defer srv2.Close()

	srv1 := newTestServer(t, map[string]interface{}{
		"aeo_version": "0.1",
		"entity": map[string]interface{}{
			"id":            "https://parent.example.com/#org",
			"type":          "Organization",
			"name":          "Parent Origin",
			"canonical_url": "https://parent.example.com/",
		},
		"authority": map[string]interface{}{
			"primary_sources": []string{srv2.URL + "/something"},
		},
		"claims": []map[string]interface{}{
			{"id": "tagline", "predicate": "description", "value": "parent", "confidence": "high"},
		},
	})
	defer srv1.Close()

	c := New(Config{MaxDepth: 2, MaxFetches: 10, Concurrency: 2, FetchTimeout: time.Second})
	results, err := c.Crawl(context.Background(), srv1.URL)
	if err != nil {
		t.Fatalf("Crawl: %v", err)
	}
	if got := len(results); got < 2 {
		t.Fatalf("len(results) = %d, want >= 2 (parent + child)", got)
	}

	names := make(map[string]bool)
	for _, r := range results {
		if r.Success {
			names[r.EntityName] = true
		}
	}
	if !names["Parent Origin"] || !names["Child Origin"] {
		t.Errorf("expected both Parent and Child, got %v", names)
	}
}

func TestCrawl_404Recorded(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/aeo.json", http.NotFound)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := New(Config{MaxDepth: 0, MaxFetches: 10, Concurrency: 1, FetchTimeout: time.Second})
	results, err := c.Crawl(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Crawl: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Success {
		t.Error("expected 404 to record success=false")
	}
	if results[0].Error == "" {
		t.Error("expected error string set")
	}
}

func TestCrawl_RespectsBudget(t *testing.T) {
	srv := newTestServer(t, map[string]interface{}{
		"aeo_version": "0.1",
		"entity": map[string]interface{}{
			"id":            "https://example.com/#org",
			"type":          "Organization",
			"name":          "Solo Origin",
			"canonical_url": "https://example.com/",
		},
		"authority": map[string]interface{}{
			"primary_sources": []string{"https://example.com/"},
		},
		"claims": []map[string]interface{}{
			{"id": "tagline", "predicate": "description", "value": "test", "confidence": "high"},
		},
	})
	defer srv.Close()

	c := New(Config{MaxDepth: 5, MaxFetches: 0, Concurrency: 1, FetchTimeout: time.Second})
	results, err := c.Crawl(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("Crawl: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected zero results with MaxFetches=0, got %d", len(results))
	}
}

func TestNormalizeOrigin_StripsPath(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://example.com/", "https://example.com"},
		{"https://example.com/some/path", "https://example.com"},
		{"http://example.com:8080/", "http://example.com:8080"},
	}
	for _, c := range cases {
		got, err := normalizeOrigin(c.in)
		if err != nil {
			t.Errorf("normalizeOrigin(%q) error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("normalizeOrigin(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeOrigin_RejectsBadScheme(t *testing.T) {
	if _, err := normalizeOrigin("ftp://example.com"); err == nil {
		t.Error("expected error for ftp scheme")
	}
}

// Ensure the result records the audit mode when present.
func TestCrawl_RecordsAuditMode(t *testing.T) {
	srv := newTestServer(t, map[string]interface{}{
		"aeo_version": "0.1",
		"entity": map[string]interface{}{
			"id":            "https://example.com/#org",
			"type":          "Organization",
			"name":          "Signed Origin",
			"canonical_url": "https://example.com/",
		},
		"authority": map[string]interface{}{
			"primary_sources": []string{"https://example.com/"},
		},
		"claims": []map[string]interface{}{
			{"id": "tagline", "predicate": "description", "value": "test", "confidence": "high"},
		},
		"audit": map[string]interface{}{
			"mode":            "signature",
			"signing_key_uri": "https://example.com/key.json",
			"signature":       "eyJ...",
		},
	})
	defer srv.Close()

	c := New(Config{MaxDepth: 0, MaxFetches: 10, Concurrency: 1, FetchTimeout: time.Second})
	results, _ := c.Crawl(context.Background(), srv.URL)
	if results[0].AuditMode != string(aeo.AuditSignature) {
		t.Errorf("AuditMode = %q, want %q", results[0].AuditMode, aeo.AuditSignature)
	}
}
