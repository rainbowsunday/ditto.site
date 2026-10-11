package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "github.com/rainbowsunday/ditto.site/dna/internal/analysis"
    "github.com/rainbowsunday/ditto.site/dna/internal/ingest"
    "github.com/rainbowsunday/ditto.site/dna/internal/pipeline"
)

func main() {
    url := flag.String("url", "", "Target URL to analyze")
    source := flag.String("source", "", "Optional existing source capture dir (e.g. runs/example.com/.../source)")
    out := flag.String("out", "./output", "Output root directory")
    flag.Parse()

    if *url == "" && *source == "" {
        log.Fatal("usage: dna --url <https://example.com> [--source <source-dir>] [--out ./output]")
    }

    if *url == "" {
        *url = "https://example.com"
    }

    if *source == "" {
        *source = ""
    }

    siteName := sanitizeHost(*url)
    rootDir := filepath.Join(*out, siteName)
    cloneDir := filepath.Join(rootDir, ".clone")

    if err := os.MkdirAll(cloneDir, 0o755); err != nil {
        log.Fatalf("create clone dir: %v", err)
    }

    capDir, err := ingest.ResolveSourceDir(*source, *url, rootDir)
    if err != nil {
        log.Fatalf("resolve capture source: %v", err)
    }

    p := pipeline.NewGenerator(cloneDir, *url, capDir)
    if err := p.Run(); err != nil {
        log.Fatalf("generate DNA: %v", err)
    }

    fmt.Printf("✓ DNA generated at %s\n", cloneDir)
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
    host = strings.Map(func(r rune) rune {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
            return r
        }
        return '-'
    }, host)
    host = strings.Trim(host, ".-")
    if host == "" {
        return "site"
    }
    return host
}

func init() {
    // Ensure deterministic ordering for help output.
    sort.Strings([]string{})
}

var _ = analysis.Section{}
var _ = ingest.ResolveSourceDir
var _ = pipeline.NewGenerator
