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
	"io"
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
	return runWithArgs(os.Args[1:], os.Stdout, os.Stderr)
}

func runWithArgs(args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("aeo-crawler", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var (
		seed        = flags.String("seed", "", "seed origin URL (required)")
		depth       = flags.Int("depth", 2, "maximum graph depth from the seed")
		maxFetches  = flags.Int("max-fetches", 100, "maximum total fetches across the run")
		concurrency = flags.Int("concurrency", 4, "maximum in-flight HTTP requests")
		timeoutSec  = flags.Int("timeout", 10, "per-request timeout in seconds")
		format      = flags.String("format", "summary", "JSONL output: summary or graph")
	)
	if err := flags.Parse(args); err != nil {
		return err
	}

	if *seed == "" {
		flags.Usage()
		return fmt.Errorf("--seed is required")
	}
	if *format != "summary" && *format != "graph" {
		return fmt.Errorf("--format must be summary or graph")
	}

	cfg := crawler.Config{
		MaxDepth:     *depth,
		MaxFetches:   *maxFetches,
		Concurrency:  *concurrency,
		FetchTimeout: time.Duration(*timeoutSec) * time.Second,
		GraphOutput:  *format == "graph",
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	c := crawler.New(cfg)
	results, err := c.Crawl(ctx, *seed)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(stdout)
	if *format == "graph" {
		// Validate all successful rows before writing to stdout so a malformed
		// declaration cannot leave a plausible but incomplete graph file.
		nodes := make([]crawler.GraphNode, 0, len(results))
		for _, r := range results {
			if !r.Success {
				continue
			}
			node, ok := r.AsGraphNode()
			if !ok {
				return fmt.Errorf("successful crawl at %s has no graph node", r.Origin)
			}
			nodes = append(nodes, node)
		}
		if len(nodes) == 0 {
			return fmt.Errorf("graph export has no successful AEO declarations")
		}
		for _, node := range nodes {
			if err := enc.Encode(node); err != nil {
				return err
			}
		}
	} else {
		for _, r := range results {
			if err := enc.Encode(r); err != nil {
				return err
			}
		}
	}

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}
	fmt.Fprintf(
		stderr,
		"\naeo-crawler: %d origins attempted, %d AEO declarations found\n",
		len(results),
		successCount,
	)
	return nil
}
