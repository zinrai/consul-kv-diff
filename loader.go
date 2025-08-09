package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LocalKVPair represents the structure of local JSON (consul kv export/import format)
type LocalKVPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ConsulKVPair represents the structure of Consul API response
type ConsulKVPair struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

// LoadLocalKV loads KV pairs from a local JSON file
func LoadLocalKV(filepath string) (map[string]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var pairs []LocalKVPair
	if err := json.NewDecoder(file).Decode(&pairs); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	kvMap := make(map[string]string)
	for _, pair := range pairs {
		kvMap[pair.Key] = pair.Value
	}

	return kvMap, nil
}

// LoadConsulKV retrieves KV pairs from Consul API
func LoadConsulKV(addr, datacenter, prefix string) (map[string]string, error) {
	// Build URL
	url := fmt.Sprintf("%s/v1/kv/%s?recurse", addr, prefix)
	if prefix == "" {
		url = fmt.Sprintf("%s/v1/kv/?recurse", addr)
	}

	// Add datacenter parameter if specified
	if datacenter != "" && datacenter != "dc1" {
		url = fmt.Sprintf("%s&dc=%s", url, datacenter)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Return empty map for 404 (keys don't exist)
		return make(map[string]string), nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("consul returned status %d: %s", resp.StatusCode, body)
	}

	var pairs []ConsulKVPair
	if err := json.NewDecoder(resp.Body).Decode(&pairs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	kvMap := make(map[string]string)
	for _, pair := range pairs {
		kvMap[pair.Key] = pair.Value
	}

	return kvMap, nil
}
