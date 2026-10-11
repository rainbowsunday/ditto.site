package ingest

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "strings"
)

// Snapshot is the minimal shape used by theDNA inference pass.
type Snapshot struct {
    Doc   map[string]any `json:"doc"`
    Root  map[string]any `json:"root"`
    Title string
    URL   string
}

// ResolveSourceDir resolves the existing capture source.
// If --source is supplied, it is used directly.
// Otherwise it looks for a standard generated source under <out>/<site>/.clone or runs/<site>/latest/source.
func ResolveSourceDir(sourceArg string, url string, outRoot string) (string, error) {
    if sourceArg != "" {
        if _, err := os.Stat(sourceArg); err == nil {
            return sourceArg, nil
        }
        return "", fmt.Errorf("source dir does not exist: %s", sourceArg)
    }

    // Common generated layout: <out>/<site>/.clone/source or <out>/<site>/source.
    candidates := []string{
        filepath.Join(outRoot, ".clone", "source"),
        filepath.Join(outRoot, "source"),
        filepath.Join("runs", sanitizeHost(url), "latest", "source"),
        filepath.Join("runs", sanitizeHost(url), "source"),
    }
    for _, c := range candidates {
        if _, err := os.Stat(c); err == nil {
            return c, nil
        }
    }
    return "", fmt.Errorf("no source capture found; provide --source or run the existing compiler capture first")
}

// LoadSnapshot reads the canonical capture file from sourceDir/capture.
func LoadSnapshot(sourceDir string) (*Snapshot, error) {
    captureDir := filepath.Join(sourceDir, "capture")
    entries, err := os.ReadDir(captureDir)
    if err != nil {
        return nil, fmt.Errorf("read capture dir: %w", err)
    }
    files := make([]string, 0, len(entries))
    for _, e := range entries {
        if !e.IsDir() && strings.HasPrefix(e.Name(), "dom-") && strings.HasSuffix(e.Name(), ".json") {
            files = append(files, filepath.Join(captureDir, e.Name()))
        }
    }
    if len(files) == 0 {
        return nil, fmt.Errorf("no dom capture files found under %s", captureDir)
    }
    sort.Strings(files)

    // Prefer the canonical viewport if present.
    selected := files[0]
    for _, f := range files {
        if strings.Contains(filepath.Base(f), "1280") {
            selected = f
            break
        }
    }

    b, err := os.ReadFile(selected)
    if err != nil {
        return nil, fmt.Errorf("read capture file %s: %w", selected, err)
    }

    var snap Snapshot
    if err := json.Unmarshal(b, &snap); err != nil {
        return nil, fmt.Errorf("unmarshal capture JSON: %w", err)
    }
    // Store a normalized title/url if missing.
    if snap.Doc == nil {
        snap.Doc = map[string]any{}
    }
    if snap.Title == "" {
        if v, ok := snap.Doc["title"]; ok {
            snap.Title, _ = v.(string)
        }
    }
    if snap.URL == "" {
        if v, ok := snap.Doc["url"]; ok {
            snap.URL, _ = v.(string)
        }
    }
    if snap.Root == nil {
        snap.Root = map[string]any{}
    }
    return &snap, nil
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
