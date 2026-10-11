package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/rainbowsunday/ditto.site/dna/internal/ingest"
	"github.com/rainbowsunday/ditto.site/dna/internal/pipeline"
)

func main() {
	url := flag.String("url", "", "Target URL")
	source := flag.String("source", "", "Source capture directory")
	out := flag.String("out", "./output", "Output root directory")
	flag.Parse()

	if *url == "" && *source == "" {
		flag.Usage()
		log.Fatal("error: --url or --source required")
	}

	// Resolve source directory
	sourceDir := *source
	if sourceDir == "" {
		candidates := []string{
			filepath.Join(*out, sanitizeHost(*url), ".clone", "source"),
			filepath.Join("runs", sanitizeHost(*url), "latest", "source"),
			filepath.Join("runs", sanitizeHost(*url), "source"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				sourceDir = c
				break
			}
		}
		if sourceDir == "" {
			log.Fatal("error: no source capture found; provide --source or run existing compiler capture first")
		}
	}

	// Verify source exists
	if _, err := os.Stat(sourceDir); err != nil {
		log.Fatalf("error: source dir not found: %s", sourceDir)
	}

	// Load URL from capture if not provided
	if *url == "" {
		capResult, err := ingest.LoadCaptureResult(sourceDir)
		if err != nil {
			log.Fatalf("error: load URL from source: %v", err)
		}
		*url = capResult.URL
	}

	siteName := sanitizeHost(*url)
	rootDir := filepath.Join(*out, siteName)
	cloneDir := filepath.Join(rootDir, ".clone")

	// Create clone directory
	if err := os.MkdirAll(cloneDir, 0o755); err != nil {
		log.Fatalf("error: mkdir %s: %v", cloneDir, err)
	}

	// Run generator
	gen := pipeline.NewGenerator(cloneDir, sourceDir, *url)
	if err := gen.Run(); err != nil {
		log.Fatalf("error: generate DNA: %v", err)
	}

	fmt.Printf("✓ DNA generated at %s/generated\n", cloneDir)
}

func sanitizeHost(raw string) string {
	host := strings.TrimSpace(raw)
	if host == "" {
		return "site"
	}

	if !strings.Contains(host, "://") {
		host = "https://" + host
	}

	u := strings.TrimPrefix(host, "https://")
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimSuffix(u, "/")
	u = strings.TrimSpace(u)

	if u == "" {
		return "site"
	}

	parts := strings.Split(u, "/")
	host = parts[0]
	host = strings.TrimPrefix(host, "www.")
	host = strings.ToLower(host)

	for _, r := range host {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
			continue
		}
		host = strings.ReplaceAll(host, string(r), "-")
	}

	host = strings.Trim(host, ".-")
	if host == "" {
		return "site"
	}

	return host
}
