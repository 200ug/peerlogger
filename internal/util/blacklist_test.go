package util_test

import (
	"testing"

	"github.com/200ug/peerlogger/internal/util"
)

func TestNewBlacklist(t *testing.T) {
	tests := []struct {
		name            string
		ipBlacklist     []string
		pubkeyBlacklist []string
		expectedIPs     int
		expectedCIDRs   int
		expectedPubkeys int
	}{
		{
			name:            "empty lists",
			ipBlacklist:     []string{},
			pubkeyBlacklist: []string{},
			expectedIPs:     0,
			expectedCIDRs:   0,
			expectedPubkeys: 0,
		},
		{
			name:            "valid ips and pubkeys",
			ipBlacklist:     []string{"192.168.1.1", "10.0.0.1"},
			pubkeyBlacklist: []string{"pubkey1", "pubkey2"},
			expectedIPs:     2,
			expectedCIDRs:   0,
			expectedPubkeys: 2,
		},
		{
			name:            "mixed ips and cidrs",
			ipBlacklist:     []string{"192.168.1.1", "10.0.0.0/24", "172.16.0.1"},
			pubkeyBlacklist: []string{"pubkey1"},
			expectedIPs:     2,
			expectedCIDRs:   1,
			expectedPubkeys: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bl := util.NewBlacklist(tt.ipBlacklist, tt.pubkeyBlacklist)

			ips, cidrs, pubkeys := bl.GetStats()
			if ips != tt.expectedIPs {
				t.Errorf("Expected %d IPs, got %d", tt.expectedIPs, ips)
			}
			if cidrs != tt.expectedCIDRs {
				t.Errorf("Expected %d CIDRs, got %d", tt.expectedCIDRs, cidrs)
			}
			if pubkeys != tt.expectedPubkeys {
				t.Errorf("Expected %d pubkeys, got %d", tt.expectedPubkeys, pubkeys)
			}
		})
	}
}

func TestBlacklist_IsIPBlacklisted(t *testing.T) {
	bl := util.NewBlacklist([]string{
		"192.168.1.1",
		"10.0.0.0/24",
		"172.16.0.0/16",
	}, []string{"pubkey1"})
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "exact ip match",
			ip:       "192.168.1.1",
			expected: true,
		},
		{
			name:     "ip in cidr range",
			ip:       "10.0.0.5",
			expected: true,
		},
		{
			name:     "ip in larger cidr range",
			ip:       "172.16.1.100",
			expected: true,
		},
		{
			name:     "ip not blacklisted",
			ip:       "8.8.8.8",
			expected: false,
		},
		{
			name:     "invalid ip format",
			ip:       "invalid-ip",
			expected: false,
		},
		{
			name:     "ip outside cidr range",
			ip:       "10.1.0.1",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bl.IsIPBlacklisted(tt.ip)
			if result != tt.expected {
				t.Errorf("IsIPBlacklisted(%s) = %v, expected %v", tt.ip, result, tt.expected)
			}
		})
	}
}

func TestBlacklist_IsPubkeyBlacklisted(t *testing.T) {
	bl := util.NewBlacklist([]string{"192.168.1.1"}, []string{
		"pubkey1",
		"pubkey2",
		"0x1234567890abcdef",
	})
	tests := []struct {
		name     string
		pubkey   string
		expected bool
	}{
		{
			name:     "pubkey in blacklist",
			pubkey:   "pubkey1",
			expected: true,
		},
		{
			name:     "another pubkey in blacklist",
			pubkey:   "pubkey2",
			expected: true,
		},
		{
			name:     "hex pubkey in blacklist",
			pubkey:   "0x1234567890abcdef",
			expected: true,
		},
		{
			name:     "pubkey not in blacklist",
			pubkey:   "pubkey3",
			expected: false,
		},
		{
			name:     "empty pubkey",
			pubkey:   "",
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bl.IsPubkeyBlacklisted(tt.pubkey)
			if result != tt.expected {
				t.Errorf("IsPubkeyBlacklisted(%s) = %v, expected %v", tt.pubkey, result, tt.expected)
			}
		})
	}
}

func TestBlacklist_Reload(t *testing.T) {
	bl := util.NewBlacklist([]string{"192.168.1.1"}, []string{"pubkey1"})
	ips, cidrs, pubkeys := bl.GetStats()
	if ips != 1 || cidrs != 0 || pubkeys != 1 {
		t.Errorf("Initial state incorrect: IPs=%d, CIDRs=%d, Pubkeys=%d", ips, cidrs, pubkeys)
	}
	bl.Reload([]string{"10.0.0.0/8", "172.16.1.1"}, []string{"pubkey2", "pubkey3", "pubkey4"})
	ips, cidrs, pubkeys = bl.GetStats()
	if ips != 1 || cidrs != 1 || pubkeys != 3 {
		t.Errorf("Reloaded state incorrect: IPs=%d, CIDRs=%d, Pubkeys=%d", ips, cidrs, pubkeys)
	}
	if bl.IsIPBlacklisted("192.168.1.1") {
		t.Error("Old IP should not be blacklisted after reload")
	}
	if !bl.IsIPBlacklisted("172.16.1.1") {
		t.Error("New IP should be blacklisted after reload")
	}
	if bl.IsPubkeyBlacklisted("pubkey1") {
		t.Error("Old pubkey should not be blacklisted after reload")
	}
	if !bl.IsPubkeyBlacklisted("pubkey2") {
		t.Error("New pubkey should be blacklisted after reload")
	}
}

func TestBlacklist_EmptyAndWhitespaceHandling(t *testing.T) {
	bl := util.NewBlacklist([]string{"", "  ", "192.168.1.1", " 10.0.0.1 "}, []string{"", "  ", "pubkey1", " pubkey2 "})
	ips, cidrs, pubkeys := bl.GetStats()
	if ips != 2 || cidrs != 0 || pubkeys != 2 {
		t.Errorf("Empty/whitespace handling incorrect: IPs=%d, CIDRs=%d, Pubkeys=%d", ips, cidrs, pubkeys)
	}
	if !bl.IsIPBlacklisted("192.168.1.1") {
		t.Error("Trimmed IP should be blacklisted")
	}
	if !bl.IsIPBlacklisted("10.0.0.1") {
		t.Error("Trimmed IP should be blacklisted")
	}
	if !bl.IsPubkeyBlacklisted("pubkey1") {
		t.Error("Trimmed pubkey should be blacklisted")
	}
	if !bl.IsPubkeyBlacklisted("pubkey2") {
		t.Error("Trimmed pubkey should be blacklisted")
	}
}

func TestBlacklist_InvalidCIDRAndIP(t *testing.T) {
	bl := util.NewBlacklist([]string{
		"invalid-cidr/24",
		"999.999.999.999",
		"not-an-ip",
		"192.168.1.1", // valid one
	}, []string{})
	ips, cidrs, pubkeys := bl.GetStats()
	if ips != 1 || cidrs != 0 || pubkeys != 0 {
		t.Errorf("Invalid IP/CIDR handling incorrect: IPs=%d, CIDRs=%d, Pubkeys=%d", ips, cidrs, pubkeys)
	}
	if !bl.IsIPBlacklisted("192.168.1.1") {
		t.Error("Valid IP should be blacklisted")
	}
}
