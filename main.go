package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Injected at build time by goreleaser via -ldflags -X.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type Config struct {
	LocalFile    string
	ConsulAddr   string
	Datacenter   string
	ConsulPrefix string
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: consul-kv-diff -local <file> [options]\n\n")
		fmt.Fprintf(os.Stderr, "consul-kv-diff compares local KV JSON with Consul KV store.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
}

func parseFlags() *Config {
	cfg := &Config{}

	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Print version information and exit")
	flag.StringVar(&cfg.LocalFile, "local", "", "Path to local KV JSON file (required)")
	flag.StringVar(&cfg.ConsulAddr, "consul-addr", "http://127.0.0.1:8500", "Consul HTTP API address")
	flag.StringVar(&cfg.Datacenter, "datacenter", "", "Consul datacenter (default: agent's own datacenter)")
	flag.StringVar(&cfg.ConsulPrefix, "prefix", "", "KV prefix to compare (optional)")

	flag.Parse()

	if showVersion {
		fmt.Printf("consul-kv-diff %s (commit %s, built %s)\n", version, commit, date)
		os.Exit(0)
	}

	if cfg.LocalFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -local flag is required\n\n")
		flag.Usage()
		os.Exit(2)
	}

	return cfg
}

func main() {
	cfg := parseFlags()

	// Load local KV
	localKV, err := LoadLocalKV(cfg.LocalFile)
	if err != nil {
		exitWithError("loading local KV", err)
	}

	// Retrieve KV from Consul
	consulKV, err := LoadConsulKV(cfg.ConsulAddr, cfg.Datacenter, cfg.ConsulPrefix)
	if err != nil {
		exitWithError("loading Consul KV", err)
	}

	// Filter by prefix if needed
	if cfg.ConsulPrefix != "" {
		localKV = filterByPrefix(localKV, cfg.ConsulPrefix)
	}

	// Detect differences
	result := Compare(localKV, consulKV)

	// Output results
	printResult(result)

	// Exit codes: 0 no differences, 1 differences found, 2 error (see exitWithError).
	if result.HasDifferences() {
		os.Exit(1)
	}
}

func filterByPrefix(kv map[string]string, prefix string) map[string]string {
	filtered := make(map[string]string)
	for key, value := range kv {
		if strings.HasPrefix(key, prefix) {
			filtered[key] = value
		}
	}
	return filtered
}

func printResult(result *DiffResult) {
	if !result.HasDifferences() {
		fmt.Println("No differences found.")
		return
	}

	if len(result.Modified) > 0 {
		fmt.Println("=== Modified Keys ===")
		// Sort by key for output
		sort.Slice(result.Modified, func(i, j int) bool {
			return result.Modified[i].Key < result.Modified[j].Key
		})
		for _, m := range result.Modified {
			fmt.Printf("Key: %s\n", m.Key)
			fmt.Printf("  Local:  %s\n", displayValue(m.OldValue))
			fmt.Printf("  Consul: %s\n", displayValue(m.NewValue))
			fmt.Println()
		}
	}

	if len(result.OnlyInConsul) > 0 {
		fmt.Println("=== Keys Only in Consul ===")
		// Sort by key for output
		sort.Strings(result.OnlyInConsul)
		for _, key := range result.OnlyInConsul {
			fmt.Printf("- %s\n", key)
		}
		fmt.Println()
	}

	if len(result.OnlyInLocal) > 0 {
		fmt.Println("=== Keys Only in Local ===")
		// Sort by key for output
		sort.Strings(result.OnlyInLocal)
		for _, key := range result.OnlyInLocal {
			fmt.Printf("- %s\n", key)
		}
	}
}

// displayValue decodes a base64-encoded KV value for human-readable output.
// Values are compared in their base64 form for exactness, but rendering the
// decoded text makes drift output actionable. Binary or non-printable values
// fall back to the base64 form so the terminal is never corrupted.
func displayValue(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || !isPrintable(decoded) {
		return encoded
	}
	return string(decoded)
}

func isPrintable(b []byte) bool {
	if !utf8.Valid(b) {
		return false
	}
	for _, r := range string(b) {
		if r == '\t' || r == '\n' || r == '\r' {
			continue
		}
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func exitWithError(operation string, err error) {
	fmt.Fprintf(os.Stderr, "Error %s: %v\n", operation, err)
	os.Exit(2)
}
