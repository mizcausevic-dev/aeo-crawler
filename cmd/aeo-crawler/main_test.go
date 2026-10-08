package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGraphFlagEmitsExplorerRowAndSummaryDefaultStaysSmall(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/aeo.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"aeo_version":"0.1","entity":{"id":"https://example.com/#org","type":"Organization","name":"Example","canonical_url":"https://example.com/"},"authority":{"primary_sources":["https://example.com/"]},"claims":[{"id":"c1","predicate":"industry","value":"testing"}]}`))
	}))
	defer srv.Close()

	for _, format := range []string{"summary", "graph"} {
		var out, errOut bytes.Buffer
		args := []string{"--seed", srv.URL, "--depth", "0"}
		if format == "graph" {
			args = append(args, "--format", "graph")
		}
		if err := runWithArgs(args, &out, &errOut); err != nil {
			t.Fatalf("%s run: %v", format, err)
		}
		var row map[string]interface{}
		if err := json.Unmarshal(out.Bytes(), &row); err != nil {
			t.Fatalf("%s output: %v", format, err)
		}
		if !strings.Contains(errOut.String(), "1 AEO declarations found") {
			t.Fatalf("%s summary missing: %s", format, errOut.String())
		}
		if format == "summary" {
			if row["success"] != true || row["body"] != nil {
				t.Fatalf("summary output changed: %v", row)
			}
		} else {
			if row["id"] != "https://example.com/#org" || row["body"] == nil {
				t.Fatalf("invalid graph row: %v", row)
			}
		}
	}
}

func TestInvalidFormatFailsBeforeFetch(t *testing.T) {
	var out, errOut bytes.Buffer
	err := runWithArgs([]string{"--seed", "https://example.com", "--format", "unknown"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "--format must be") || out.Len() != 0 {
		t.Fatalf("unexpected result: error=%v output=%q", err, out.String())
	}
}

func TestGraphOutputFailsWhenEveryFetchFails(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	var out, errOut bytes.Buffer
	err := runWithArgs([]string{"--seed", srv.URL, "--depth", "0", "--format", "graph"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "no successful AEO declarations") || out.Len() != 0 {
		t.Fatalf("expected empty graph export failure; error=%v output=%q", err, out.String())
	}
}

func TestGraphOutputDoesNotWritePartialFileWhenADeclarationHasNoID(t *testing.T) {
	invalid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"aeo_version":"0.1","entity":{"id":"","type":"Organization","name":"Invalid","canonical_url":"https://invalid.example/"},"authority":{"primary_sources":[]},"claims":[]}`))
	}))
	defer invalid.Close()
	valid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"aeo_version":"0.1","entity":{"id":"https://valid.example/#org","type":"Organization","name":"Valid","canonical_url":"https://valid.example/"},"authority":{"primary_sources":["` + invalid.URL + `/evidence"]},"claims":[]}`))
	}))
	defer valid.Close()

	var out, errOut bytes.Buffer
	err := runWithArgs([]string{"--seed", valid.URL, "--depth", "1", "--format", "graph"}, &out, &errOut)
	if err == nil || !strings.Contains(err.Error(), "has no graph node") || out.Len() != 0 {
		t.Fatalf("expected no partial graph JSONL; error=%v output=%q", err, out.String())
	}
}
