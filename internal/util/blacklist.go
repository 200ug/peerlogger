package util

import (
	"net"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type Blacklist struct {
	ipNets  []*net.IPNet
	ips     map[string]bool
	pubkeys map[string]bool
	mu      sync.RWMutex
}

func NewBlacklist(ipBlacklist []string, pubkeyBlacklist []string) *Blacklist {
	b := &Blacklist{
		ips:     make(map[string]bool),
		pubkeys: make(map[string]bool),
	}
	// no need to acquire locks here as no one else is using the blacklist yet
	b.parseIPBlacklist(ipBlacklist)
	b.parsePubkeyBlacklist(pubkeyBlacklist)

	log.Info().Int("single_ips", len(b.ips)).Int("cidr_blocks", len(b.ipNets)).Int("pubkeys", len(b.pubkeys)).Msg("Blacklist loaded")

	return b
}

func (b *Blacklist) parseIPBlacklist(ipList []string) {
	// caller handles locking
	for _, ipStr := range ipList {
		ipStr := strings.TrimSpace(ipStr)
		if ipStr == "" {
			continue
		}

		// check if defined in cidr notation
		if strings.Contains(ipStr, "/") {
			_, ipNet, err := net.ParseCIDR(ipStr)
			if err != nil {
				log.Warn().Err(err).Str("cidr", ipStr).Msg("Invalid CIDR in blacklist, skipping")
				continue
			}
			b.ipNets = append(b.ipNets, ipNet)
		} else {
			// single ip
			ip := net.ParseIP(ipStr)
			if ip == nil {
				log.Warn().Str("ip", ipStr).Msg("Invalid IP address in blacklist, skipping")
				continue
			}
			b.ips[ipStr] = true
		}
	}
}

func (b *Blacklist) parsePubkeyBlacklist(pubkeyList []string) {
	// caller handles locking
	for _, pubkey := range pubkeyList {
		pubkey = strings.TrimSpace(pubkey)
		if pubkey == "" {
			continue
		}
		b.pubkeys[pubkey] = true
	}
}

func (b *Blacklist) Reload(ipBlacklist []string, pubkeyBlacklist []string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// clear existing lists
	b.ipNets = []*net.IPNet{}
	b.ips = make(map[string]bool)
	b.pubkeys = make(map[string]bool)

	// reparse
	b.parseIPBlacklist(ipBlacklist)
	b.parsePubkeyBlacklist(pubkeyBlacklist)

	log.Info().Int("single_ips", len(b.ips)).Int("cidr_blocks", len(b.ipNets)).Int("pubkeys", len(b.pubkeys)).Msg("Blacklist reloaded")
}

func (b *Blacklist) IsIPBlacklisted(ipStr string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// exact ip match
	if b.ips[ipStr] {
		return true
	}
	// cidr block range match
	ip := net.ParseIP(ipStr)
	if ip == nil {
		log.Warn().Str("ip", ipStr).Msg("Invalid IP address format for blacklist check")
		return false
	}
	for _, ipNet := range b.ipNets {
		if ipNet.Contains(ip) {
			return true
		}
	}

	return false
}

func (b *Blacklist) IsPubkeyBlacklisted(pubkey string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.pubkeys[pubkey]
}

func (b *Blacklist) GetStats() (ips int, ipNets int, pubkeys int) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.ips), len(b.ipNets), len(b.pubkeys)
}
