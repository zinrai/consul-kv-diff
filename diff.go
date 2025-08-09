package main

// DiffResult represents the result of difference detection
type DiffResult struct {
	Modified     []ModifiedKey
	OnlyInConsul []string
}

// ModifiedKey represents information about a modified key
type ModifiedKey struct {
	Key      string
	OldValue string // Local (YAML managed) value
	NewValue string // Value in Consul
}

// Compare detects differences between two KV maps
func Compare(local, consul map[string]string) *DiffResult {
	result := &DiffResult{
		Modified:     []ModifiedKey{},
		OnlyInConsul: []string{},
	}

	// Compare values for keys that exist locally
	for key, localValue := range local {
		if consulValue, exists := consul[key]; exists {
			if localValue != consulValue {
				result.Modified = append(result.Modified, ModifiedKey{
					Key:      key,
					OldValue: localValue,
					NewValue: consulValue,
				})
			}
		}
	}

	// Detect keys that exist only in Consul
	for key := range consul {
		if _, exists := local[key]; !exists {
			result.OnlyInConsul = append(result.OnlyInConsul, key)
		}
	}

	return result
}

// HasDifferences returns whether differences exist
func (d *DiffResult) HasDifferences() bool {
	return len(d.Modified) > 0 || len(d.OnlyInConsul) > 0
}
