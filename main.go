package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/200ug/peerlogger/internal/crawler"
	"github.com/200ug/peerlogger/internal/db"
	"github.com/200ug/peerlogger/internal/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// S3l1biBLaWxsYSBLbGFuLCB3ZSBtYWtlIGl0IGxvb2sgZWFzeQ==

const (
	defaultLogLevel = zerolog.InfoLevel
)

var (
	appCfg     *util.AppConfig
	envCfg     *util.EnvConfig
	appCfgHash [32]byte
)

func init() {
	envCfg = util.LoadEnv()
	appCfg, appCfgHash = util.LoadConfig()
	logLevel, _ := zerolog.ParseLevel(appCfg.LogLevel)
	zerolog.SetGlobalLevel(logLevel)
	log.Debug().Msg("Primary initialization done")
}

func main() {
	log.Info().Msg("Starting peerlogger")

	blacklist := crawler.NewBlacklist(appCfg.IPBlacklist, appCfg.PubkeyBlacklist)
	db := db.NewDatabase(envCfg.DBURL)
	geoIP := crawler.NewGeoIP()
	crawlerCfg := &crawler.Config{
		CrawlInterval:     appCfg.CrawlInterval,
		DiscoveryTimeout:  appCfg.DiscoveryTimeout,
		MaxParallelCrawls: appCfg.MaxParallelCrawls,
	}

	c := crawler.NewCrawler(crawlerCfg, blacklist, db, geoIP)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		c.Start(ctx)
	}()

	<-sigChan
	log.Info().Msg("Shutdown signal received, stopping crawler...")
	time.Sleep(5 * time.Second) // allow some time for graceful shutdown
	log.Info().Msg("Crawler stopped, goodbye")
}
