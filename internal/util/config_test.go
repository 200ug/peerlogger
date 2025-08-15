package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/200ug/peerlogger/internal/util"
)

func TestLoadJSONList(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name         string
		fileContent  string
		fileName     string
		key          string
		expected     []string
		shouldCreate bool
	}{
		{
			name:         "valid json with string array",
			fileContent:  `{"ip_blacklists": ["192.168.1.1", "10.0.0.1", "172.16.0.0/16"]}`,
			fileName:     "valid.json",
			key:          "ip_blacklists",
			expected:     []string{"192.168.1.1", "10.0.0.1", "172.16.0.0/16"},
			shouldCreate: true,
		},
		{
			name:         "valid json with mixed types (should skip non-strings)",
			fileContent:  `{"pubkey_blacklists": ["pubkey1", 123, "pubkey2", true, "pubkey3"]}`,
			fileName:     "mixed.json",
			key:          "pubkey_blacklists",
			expected:     []string{"pubkey1", "pubkey2", "pubkey3"},
			shouldCreate: true,
		},
		{
			name:         "empty array",
			fileContent:  `{"ip_blacklists": []}`,
			fileName:     "empty.json",
			key:          "ip_blacklists",
			expected:     []string{},
			shouldCreate: true,
		},
		{
			name:         "missing key",
			fileContent:  `{"other_key": ["value1", "value2"]}`,
			fileName:     "missing_key.json",
			key:          "ip_blacklists",
			expected:     []string{},
			shouldCreate: true,
		},
		{
			name:         "invalid json",
			fileContent:  `{"ip_blacklists": ["192.168.1.1", "10.0.0.1"`,
			fileName:     "invalid.json",
			key:          "ip_blacklists",
			expected:     []string{},
			shouldCreate: true,
		},
		{
			name:         "non-array value",
			fileContent:  `{"ip_blacklists": "not an array"}`,
			fileName:     "non_array.json",
			key:          "ip_blacklists",
			expected:     []string{},
			shouldCreate: true,
		},
		{
			name:         "file does not exist",
			fileName:     "nonexistent.json",
			key:          "ip_blacklists",
			expected:     []string{},
			shouldCreate: false,
		},
		{
			name:         "empty path",
			fileName:     "",
			key:          "ip_blacklists",
			expected:     []string{},
			shouldCreate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filePath string
			if tt.shouldCreate && tt.fileName != "" {
				filePath = filepath.Join(tmpDir, tt.fileName)
				err := os.WriteFile(filePath, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			} else if tt.fileName != "" {
				filePath = filepath.Join(tmpDir, tt.fileName)
			}

			result := util.LoadJSONList(filePath, tt.key)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected %v, got %v", tt.expected, result)
					break
				}
			}
		})
	}
}

func TestLoadJSONListRealExample(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "config_test_real")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a realistic blacklist file
	ipBlacklistContent := `{
		"ip_blacklists": [
			"127.0.0.1",
			"0.0.0.0",
			"192.168.0.0/16",
			"10.0.0.0/8",
			"172.16.0.0/12"
		]
	}`

	pubkeyBlacklistContent := `{
		"pubkey_blacklists": [
			"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"malicious_pubkey_1",
			"malicious_pubkey_2"
		]
	}`

	ipFile := filepath.Join(tmpDir, "ip_blacklist.json")
	pubkeyFile := filepath.Join(tmpDir, "pubkey_blacklist.json")

	err = os.WriteFile(ipFile, []byte(ipBlacklistContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create IP blacklist file: %v", err)
	}

	err = os.WriteFile(pubkeyFile, []byte(pubkeyBlacklistContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create pubkey blacklist file: %v", err)
	}

	// Test loading IP blacklist
	ipList := util.LoadJSONList(ipFile, "ip_blacklists")
	if len(ipList) != 5 {
		t.Errorf("Expected 5 IP blacklist entries, got %d", len(ipList))
	}

	expectedIPs := []string{"127.0.0.1", "0.0.0.0", "192.168.0.0/16", "10.0.0.0/8", "172.16.0.0/12"}
	for i, expected := range expectedIPs {
		if i >= len(ipList) || ipList[i] != expected {
			t.Errorf("Expected IP %s at index %d, got %s", expected, i, ipList[i])
		}
	}

	// Test loading pubkey blacklist
	pubkeyList := util.LoadJSONList(pubkeyFile, "pubkey_blacklists")
	if len(pubkeyList) != 3 {
		t.Errorf("Expected 3 pubkey blacklist entries, got %d", len(pubkeyList))
	}

	expectedPubkeys := []string{
		"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		"malicious_pubkey_1",
		"malicious_pubkey_2",
	}
	for i, expected := range expectedPubkeys {
		if i >= len(pubkeyList) || pubkeyList[i] != expected {
			t.Errorf("Expected pubkey %s at index %d, got %s", expected, i, pubkeyList[i])
		}
	}
}