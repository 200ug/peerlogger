package crawler

import (
	"context"
	"time"

	"github.com/200ug/peerlogger/internal/db"
	"github.com/200ug/peerlogger/internal/util"
	"github.com/rs/zerolog/log"
)

type Crawler struct {
	config *Config
	done   chan struct{}
}

type Config struct {
	ConfigHash        [32]byte
	Database          *db.Database
	NodeRepository    *db.NodeRepository
	Blacklist         *Blacklist
	GeoIP             *GeoIP
	CrawlInterval     time.Duration
	DiscoveryTimeout  time.Duration
	MaxParallelCrawls int
}

func NewCrawler(config *Config, blacklist *Blacklist, db *db.Database, geoIP *GeoIP) *Crawler {
	return &Crawler{
		config: config,
		done:   make(chan struct{}),
	}
}

func (c *Crawler) Start(ctx context.Context) {
	log.Info().Msg("Starting node crawler")
	ticker := time.NewTicker(c.config.CrawlInterval)
	defer ticker.Stop()
	defer close(c.done)

	// initial crawl
	if err := c.runRound(); err != nil {
		log.Error().Err(err).Msg("Initial crawl round failed")
	}

	for {
		select {
		case <-ticker.C:
			if err := c.runRound(); err != nil {
				log.Error().Err(err).Msg("Crawl round failed")
			}
		case <-ctx.Done():
			log.Info().Msg("Crawler stopping due to context cancellation")
			return
		}
	}
}

func (c *Crawler) runRound() error {
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
		c.config.Blacklist.Reload(newUserConfig.IPBlacklist, newUserConfig.PubkeyBlacklist)
	}
	return nil
}

// Return a channel that's closed when the crawler stops
func (c *Crawler) Done() <-chan struct{} {
	return c.done
}
