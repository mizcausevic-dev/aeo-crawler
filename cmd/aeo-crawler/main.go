// Command aeo-crawler walks an AEO graph from a seed origin and emits
// one JSON Lines record per origin attempted.
//
//	aeo-crawler --seed https://mizcausevic-dev.github.io --depth 2
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	crawler "github.com/mizcausevic-dev/aeo-crawler"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "aeo-crawler: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		seed        = flag.String("seed", "", "seed origin URL (required)")
		depth       = flag.Int("depth", 2, "maximum graph depth from the seed")
		maxFetches  = flag.Int("max-fetches", 100, "maximum total fetches across the run")
		concurrency = flag.Int("concurrency", 4, "maximum in-flight fetches")
		timeoutSec  = flag.Int("timeout", 10, "per-request timeout in seconds")
	)
	flag.Parse()

	if *seed == "" {
		flag.Usage()
		return fmt.Errorf("--seed is required")
	}

	cfg := crawler.Config{
		MaxDepth:     *depth,
		MaxFetches:   *maxFetches,
		Concurrency:  *concurrency,
		FetchTimeout: time.Duration(*timeoutSec) * time.Second,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c := crawler.New(cfg)
	results, err := c.Crawl(ctx, *seed)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	for _, r := range results {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}
	fmt.Fprintf(
		os.Stderr,
		"\naeo-crawler: %d origins attempted, %d AEO declarations found\n",
		len(results),
		successCount,
	)
	return nil
}
