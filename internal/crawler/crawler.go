package crawler

import (
	"context"
	"net"
	"time"

	"github.com/200ug/peerlogger/internal/db"
	"github.com/200ug/peerlogger/internal/util"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/rs/zerolog/log"
)

type Crawler struct {
	discoveryNodes []*enode.Node
	blacklist      *Blacklist
	db             *db.Database
	geoIP          *GeoIP
	config         *Config
}

type Config struct {
	ConfigHash        [32]byte
	CrawlInterval     time.Duration
	DiscoveryTimeout  time.Duration
	MaxParallelCrawls int
}

func NewCrawler(config *Config, blacklist *Blacklist, db *db.Database, geoIP *GeoIP) *Crawler {
	return &Crawler{
		config:    config,
		blacklist: blacklist,
		db:        db,
		geoIP:     geoIP,
	}
}

type NodeInfo struct {
	ID           string    `json:"id"`
	IP           net.IP    `json:"ip"`
	Port         uint16    `json:"port"`
	TCPPort      uint16    `json:"tcp_port"`
	ClientType   string    `json:"client_type"`
	CountryCode  string    `json:"country_code"`
	City         string    `json:"city"`
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	LastSeen     time.Time `json:"last_seen"`
	CrawledAt    time.Time `json:"crawled_at"`
	Capabilities []p2p.Cap `json:"capabilities"`
}

func (c *Crawler) Start(ctx context.Context) {
	log.Info().Msg("Starting node crawler")

	// init discovery nodes (bootnodes)
	c.initializeDiscoveryNodes()

	ticker := time.NewTicker(c.config.CrawlInterval)
	defer ticker.Stop()

	// initial crawl
	c.runRound()

	for {
		select {
		case <-ticker.C:
			c.runRound()
		case <-ctx.Done():
			log.Info().Msg("Stopping node crawler (ctx)")
			return
		}
	}
}

func (c *Crawler) initializeDiscoveryNodes() {
	// todo: init bootnodes
	log.Debug().Msg("'initializeDiscoveryNodes' called, not implemented yet")
}

func (c *Crawler) runRound() {
	log.Info().Msg("Starting crawl round")

	// only reload the config (and related rules) if there's been changes
	newUserConfig, newHash := util.LoadConfig()
	if newHash != c.config.ConfigHash {
		c.config = &Config{
			ConfigHash:        newHash,
			CrawlInterval:     newUserConfig.CrawlInterval,
			DiscoveryTimeout:  newUserConfig.DiscoveryTimeout,
			MaxParallelCrawls: newUserConfig.MaxParallelCrawls,
		}
		c.blacklist.Reload(newUserConfig.IPBlacklist, newUserConfig.PubkeyBlacklist)
	}
}
