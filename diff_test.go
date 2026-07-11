package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestCompare_CoreScenarios(t *testing.T) {
	tests := []struct {
		name   string
		local  map[string]string
		consul map[string]string
		want   DiffResult
	}{
		{
			name:   "Detects value changes",
			local:  map[string]string{"app/host": "bG9jYWxob3N0", "app/port": "ODA4MA=="},
			consul: map[string]string{"app/host": "bG9jYWxob3N0", "app/port": "OTA5MA=="},
			want: DiffResult{
				Modified: []ModifiedKey{
					{Key: "app/port", OldValue: "ODA4MA==", NewValue: "OTA5MA=="},
				},
				OnlyInConsul: []string{},
				OnlyInLocal:  []string{},
			},
		},
		{
			name:   "Detects keys only in Consul",
			local:  map[string]string{"app/host": "bG9jYWxob3N0"},
			consul: map[string]string{"app/host": "bG9jYWxob3N0", "app/debug": "dHJ1ZQ=="},
			want: DiffResult{
				Modified:     []ModifiedKey{},
				OnlyInConsul: []string{"app/debug"},
				OnlyInLocal:  []string{},
			},
		},
		{
			name:   "Detects keys only in local",
			local:  map[string]string{"app/host": "bG9jYWxob3N0", "app/gone": "dmFsdWU="},
			consul: map[string]string{"app/host": "bG9jYWxob3N0"},
			want: DiffResult{
				Modified:     []ModifiedKey{},
				OnlyInConsul: []string{},
				OnlyInLocal:  []string{"app/gone"},
			},
		},
		{
			name:   "No differences",
			local:  map[string]string{"app/host": "bG9jYWxob3N0", "app/port": "ODA4MA=="},
			consul: map[string]string{"app/host": "bG9jYWxob3N0", "app/port": "ODA4MA=="},
			want: DiffResult{
				Modified:     []ModifiedKey{},
				OnlyInConsul: []string{},
				OnlyInLocal:  []string{},
			},
		},
		{
			name:   "Multiple changes across all categories",
			local:  map[string]string{"key1": "dmFsdWUx", "key2": "dmFsdWUy", "key4": "dmFsdWU0"},
			consul: map[string]string{"key1": "bmV3MQ==", "key2": "dmFsdWUy", "key3": "dmFsdWUz"},
			want: DiffResult{
				Modified: []ModifiedKey{
					{Key: "key1", OldValue: "dmFsdWUx", NewValue: "bmV3MQ=="},
				},
				OnlyInConsul: []string{"key3"},
				OnlyInLocal:  []string{"key4"},
			},
		},
		{
			name:   "Empty local KV",
			local:  map[string]string{},
			consul: map[string]string{"key1": "dmFsdWUx", "key2": "dmFsdWUy"},
			want: DiffResult{
				Modified:     []ModifiedKey{},
				OnlyInConsul: []string{"key1", "key2"},
				OnlyInLocal:  []string{},
			},
		},
		{
			name:   "Empty Consul KV",
			local:  map[string]string{"key1": "dmFsdWUx", "key2": "dmFsdWUy"},
			consul: map[string]string{},
			want: DiffResult{
				Modified:     []ModifiedKey{},
				OnlyInConsul: []string{},
				OnlyInLocal:  []string{"key1", "key2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Compare(tt.local, tt.consul)

			// Sort results (for order-independent testing)
			sort.Slice(result.Modified, func(i, j int) bool {
				return result.Modified[i].Key < result.Modified[j].Key
			})
			sort.Strings(result.OnlyInConsul)
			sort.Strings(result.OnlyInLocal)

			// Also sort expected values
			sort.Slice(tt.want.Modified, func(i, j int) bool {
				return tt.want.Modified[i].Key < tt.want.Modified[j].Key
			})
			sort.Strings(tt.want.OnlyInConsul)
			sort.Strings(tt.want.OnlyInLocal)

			if !reflect.DeepEqual(result.Modified, tt.want.Modified) {
				t.Errorf("Modified keys mismatch\ngot:  %+v\nwant: %+v", result.Modified, tt.want.Modified)
			}

			if !reflect.DeepEqual(result.OnlyInConsul, tt.want.OnlyInConsul) {
				t.Errorf("OnlyInConsul keys mismatch\ngot:  %+v\nwant: %+v", result.OnlyInConsul, tt.want.OnlyInConsul)
			}

			if !reflect.DeepEqual(result.OnlyInLocal, tt.want.OnlyInLocal) {
				t.Errorf("OnlyInLocal keys mismatch\ngot:  %+v\nwant: %+v", result.OnlyInLocal, tt.want.OnlyInLocal)
			}
		})
	}
}

func TestDiffResult_HasDifferences(t *testing.T) {
	tests := []struct {
		name   string
		result DiffResult
		want   bool
	}{
		{
			name: "Has modifications",
			result: DiffResult{
				Modified: []ModifiedKey{{Key: "test", OldValue: "old", NewValue: "new"}},
			},
			want: true,
		},
		{
			name: "Has Consul-only keys",
			result: DiffResult{
				OnlyInConsul: []string{"test"},
			},
			want: true,
		},
		{
			name: "Has local-only keys",
			result: DiffResult{
				OnlyInLocal: []string{"test"},
			},
			want: true,
		},
		{
			name: "Has both",
			result: DiffResult{
				Modified:     []ModifiedKey{{Key: "test", OldValue: "old", NewValue: "new"}},
				OnlyInConsul: []string{"test2"},
			},
			want: true,
		},
		{
			name:   "No differences",
			result: DiffResult{},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.HasDifferences(); got != tt.want {
				t.Errorf("HasDifferences() = %v, want %v", got, tt.want)
			}
		})
	}
}
