package db_test

import (
	"testing"
	"time"

	"github.com/200ug/peerlogger/internal/db"
	_ "github.com/lib/pq"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		config      db.Config
		wantErr     bool
		errContains string
	}{
		{
			name: "invalid database URL",
			config: db.Config{
				URL:             "invalid-url",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: time.Hour,
			},
			wantErr:     true,
			errContains: "failed to ping database",
		},
		{
			name: "empty URL",
			config: db.Config{
				URL:             "",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: time.Hour,
			},
			wantErr:     true,
			errContains: "failed to ping database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, err := db.New(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					errString := err.Error()
					if !containsSubstring(errString, tt.errContains) {
						t.Errorf("Expected error to contain '%s', got '%s'", tt.errContains, errString)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if database == nil {
				t.Error("Expected non-nil database, got nil")
				return
			}

			database.Close()
		})
	}
}

// Helper function to check if a string contains a substring.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
