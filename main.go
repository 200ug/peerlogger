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

var (
	appCfg     *util.AppConfig
	envCfg     *util.EnvConfig
	appCfgHash [32]byte
)

func init() {
	envCfg = util.LoadEnv()
	appCfg, appCfgHash = util.LoadConfig()
	logLevel, _ := zerolog.ParseLevel(appCfg.LogLevel) // should already be verified
	zerolog.SetGlobalLevel(logLevel)
	log.Debug().Msg("Primary initialization done")
}

func main() {
	log.Info().Msg("Starting peerlogger")

	blacklist := crawler.NewBlacklist(appCfg.IPBlacklist, appCfg.PubkeyBlacklist)

	dbConfig := db.Config{
		URL:             envCfg.DBURL,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
	database, err := db.New(dbConfig)
	if err != nil {
		log.Fatal().Err(err).Str("db_url", envCfg.DBURL).Msg("Failed to connect to database")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := database.HealthCheck(ctx); err != nil {
		database.Close()
		log.Fatal().Err(err).Str("db_url", envCfg.DBURL).Msg("Database health check failed")
	}
	log.Info().Msg("Database connection established successfully")
	nodeRepo := db.NewNodeRepository(database)

	geoIP := crawler.NewGeoIP()

	crawlerCfg := &crawler.Config{
		ConfigHash:        appCfgHash,
		Database:          database,
		NodeRepository:    nodeRepo,
		Blacklist:         blacklist,
		GeoIP:             geoIP,
		CrawlInterval:     appCfg.CrawlInterval,
		DiscoveryTimeout:  appCfg.DiscoveryTimeout,
		MaxParallelCrawls: appCfg.MaxParallelCrawls,
	}

	c := crawler.NewCrawler(crawlerCfg, blacklist, database, geoIP)
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
