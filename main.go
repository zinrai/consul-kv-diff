package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
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
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  consul-kv-diff -local export.json\n")
		fmt.Fprintf(os.Stderr, "  consul-kv-diff -local export.json -consul-addr http://consul:8500\n")
		fmt.Fprintf(os.Stderr, "  consul-kv-diff -local export.json -datacenter us-east-1\n")
		fmt.Fprintf(os.Stderr, "  consul-kv-diff -local export.json -prefix app/production\n")
	}
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.LocalFile, "local", "", "Path to local KV JSON file (required)")
	flag.StringVar(&cfg.ConsulAddr, "consul-addr", "http://127.0.0.1:8500", "Consul HTTP API address")
	flag.StringVar(&cfg.Datacenter, "datacenter", "dc1", "Consul datacenter")
	flag.StringVar(&cfg.ConsulPrefix, "prefix", "", "KV prefix to compare (optional)")

	flag.Parse()

	if cfg.LocalFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -local flag is required\n\n")
		flag.Usage()
		os.Exit(1)
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

	// Exit with code 1 if differences found
	if result.HasDifferences() {
		os.Exit(1)
	}
}

func filterByPrefix(kv map[string]string, prefix string) map[string]string {
	filtered := make(map[string]string)
	for key, value := range kv {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			filtered[key] = value
		}
	}
	return filtered
}

func printResult(result *DiffResult) {
	if len(result.Modified) == 0 && len(result.OnlyInConsul) == 0 {
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
			fmt.Printf("  Local:  %s\n", m.OldValue)
			fmt.Printf("  Consul: %s\n", m.NewValue)
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
	}
}

func exitWithError(operation string, err error) {
	fmt.Fprintf(os.Stderr, "Error %s: %v\n", operation, err)
	os.Exit(1)
}
