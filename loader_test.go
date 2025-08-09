package main

import (
	"io"
	"os"
	"reflect"
	"testing"
)

func TestLoadLocalKV(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]string
		wantErr bool
	}{
		{
			name: "Load valid JSON",
			content: `[
				{"key": "app/host", "value": "bG9jYWxob3N0"},
				{"key": "app/port", "value": "ODA4MA=="}
			]`,
			want: map[string]string{
				"app/host": "bG9jYWxob3N0",
				"app/port": "ODA4MA==",
			},
			wantErr: false,
		},
		{
			name:    "Empty array",
			content: `[]`,
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			content: `{"invalid": "json"`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "Single entry",
			content: `[
				{"key": "single/key", "value": "c2luZ2xlIHZhbHVl"}
			]`,
			want: map[string]string{
				"single/key": "c2luZ2xlIHZhbHVl",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpfile, err := os.CreateTemp("", "test-*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			// Write test data
			if _, err := io.WriteString(tmpfile, tt.content); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// Execute test
			got, err := LoadLocalKV(tmpfile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadLocalKV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadLocalKV() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadLocalKV_FileNotFound(t *testing.T) {
	_, err := LoadLocalKV("non-existent-file.json")
	if err == nil {
		t.Error("LoadLocalKV() expected error for non-existent file, got nil")
	}
}
