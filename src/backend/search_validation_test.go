package main

import (
	"strings"
	"testing"
)

func TestValidateSearchQuery(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "valid short query",
			query:       "golang",
			expectError: false,
		},
		{
			name:        "valid medium query",
			query:       "golang web development tutorial",
			expectError: false,
		},
		{
			name:        "empty query",
			query:       "",
			expectError: false,
		},
		{
			name:        "query at max length",
			query:       strings.Repeat("a", 200),
			expectError: false,
		},
		{
			name:        "query exceeds max length",
			query:       strings.Repeat("a", 201),
			expectError: true,
		},
		{
			name:        "very long query",
			query:       strings.Repeat("a", 500),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSearchQuery(tt.query)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
