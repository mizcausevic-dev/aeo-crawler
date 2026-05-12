// Package crawler walks an AEO graph from a seed origin: fetches the
// seed's /.well-known/aeo.json, then follows every primary_source URI
// as a candidate origin to fetch in turn, up to a configurable depth
// and total fetch budget.
package crawler

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	aeo "github.com/mizcausevic-dev/aeo-sdk-go"
)

// Config controls a Crawl run.
type Config struct {
	// MaxDepth is the maximum graph distance from the seed.
	// 0 = only fetch the seed itself.
	MaxDepth int
	// MaxFetches is a global cap on total fetches, regardless of depth.
	MaxFetches int
	// Concurrency is the maximum number of in-flight fetches.
	Concurrency int
	// FetchTimeout is the per-request HTTP timeout.
	FetchTimeout time.Duration
}

// DefaultConfig returns a sensible default crawl configuration.
func DefaultConfig() Config {
	return Config{
		MaxDepth:     2,
		MaxFetches:   100,
		Concurrency:  4,
		FetchTimeout: 10 * time.Second,
	}
}

// Result is one entry in the crawl output.
type Result struct {
	Origin      string  `json:"origin"`
	Depth       int     `json:"depth"`
	Success     bool    `json:"success"`
	EntityName  string  `json:"entity_name,omitempty"`
	EntityType  string  `json:"entity_type,omitempty"`
	ClaimsCount int     `json:"claims_count,omitempty"`
	AuditMode   string  `json:"audit_mode,omitempty"`
	Error       string  `json:"error,omitempty"`
	FetchedAt   string  `json:"fetched_at"`
}

// Crawler walks an AEO declaration graph.
type Crawler struct {
	cfg    Config
	client *aeo.Client

	mu        sync.Mutex
	visited   map[string]bool
	results   []Result
	fetchBudget int
}

// New constructs a Crawler with the given config.
func New(cfg Config) *Crawler {
	httpClient := aeo.DefaultClient()
	httpClient.HTTPClient.Timeout = cfg.FetchTimeout
	return &Crawler{
		cfg:         cfg,
		client:      httpClient,
		visited:     make(map[string]bool),
		fetchBudget: cfg.MaxFetches,
	}
}

// Crawl walks the AEO graph starting from seed and returns one Result
// per origin attempted. The context cancels the entire crawl.
func (c *Crawler) Crawl(ctx context.Context, seed string) ([]Result, error) {
	seedOrigin, err := normalizeOrigin(seed)
	if err != nil {
		return nil, fmt.Errorf("crawler: invalid seed %q: %w", seed, err)
	}

	// BFS frontier per depth.
	frontier := []string{seedOrigin}
	for depth := 0; depth <= c.cfg.MaxDepth && len(frontier) > 0; depth++ {
		next := c.fetchFrontier(ctx, frontier, depth)
		frontier = next
	}

	return c.results, nil
}

func (c *Crawler) fetchFrontier(ctx context.Context, origins []string, depth int) []string {
	sem := make(chan struct{}, c.cfg.Concurrency)
	var wg sync.WaitGroup
	var nextMu sync.Mutex
	next := make([]string, 0, len(origins))

	for _, origin := range origins {
		if !c.reserveFetch(origin) {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(origin string) {
			defer wg.Done()
			defer func() { <-sem }()

			res, follow := c.fetchOne(ctx, origin, depth)
			c.recordResult(res)

			if len(follow) > 0 {
				nextMu.Lock()
				next = append(next, follow...)
				nextMu.Unlock()
			}
		}(origin)
	}

	wg.Wait()
	return next
}

func (c *Crawler) reserveFetch(origin string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.visited[origin] {
		return false
	}
	if c.fetchBudget <= 0 {
		return false
	}
	c.visited[origin] = true
	c.fetchBudget--
	return true
}

func (c *Crawler) recordResult(r Result) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = append(c.results, r)
}

func (c *Crawler) fetchOne(ctx context.Context, origin string, depth int) (Result, []string) {
	r := Result{
		Origin:    origin,
		Depth:     depth,
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
	}

	doc, err := c.client.FetchWellKnown(ctx, origin)
	if err != nil {
		r.Success = false
		r.Error = err.Error()
		var httpErr *aeo.HTTPStatusError
		if errors.As(err, &httpErr) {
			r.Error = fmt.Sprintf("HTTP %d", httpErr.Status)
		}
		return r, nil
	}

	r.Success = true
	r.EntityName = doc.Entity.Name
	r.EntityType = string(doc.Entity.Type)
	r.ClaimsCount = len(doc.Claims)
	if doc.Audit != nil {
		r.AuditMode = string(doc.Audit.Mode)
	}

	follow := candidateOrigins(doc)
	return r, follow
}

// candidateOrigins extracts unique origins from the document's
// authority.primary_sources to attempt at the next depth.
func candidateOrigins(doc *aeo.Document) []string {
	seen := make(map[string]bool)
	out := make([]string, 0)
	for _, src := range doc.Authority.PrimarySources {
		origin, err := normalizeOrigin(src)
		if err != nil {
			continue
		}
		if !seen[origin] {
			seen[origin] = true
			out = append(out, origin)
		}
	}
	return out
}

// normalizeOrigin parses a URI and returns "scheme://host[:port]".
func normalizeOrigin(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	if u.Host == "" {
		return "", errors.New("missing host")
	}
	return u.Scheme + "://" + u.Host, nil
}
