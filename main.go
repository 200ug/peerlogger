package main

import (
	"context"
	"fmt"
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

func initDB() (*db.NodeRepository, error) {
	dbConfig := db.Config{
		URL:             envCfg.DBURL,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
	database, err := db.New(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("initialization failed: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := database.HealthCheck(ctx); err != nil {
		database.Close()
		return nil, fmt.Errorf("healthcheck failed: %w", err)
	}
	nodeRepo := db.NewNodeRepository(database)

	log.Info().Msg("Database connection established successfully")

	return nodeRepo, nil
}

func initGeoIP() (*crawler.GeoIP, error) {
	if appCfg.GeoIPCityDBPath == "" && appCfg.GeoIPASNDBPath == "" {
		log.Info().Msg("No GeoIP database paths configured, skipping GeoIP initialization")
		return nil, nil
	}
	provider, err := crawler.NewGeoIP(appCfg.GeoIPCityDBPath, appCfg.GeoIPASNDBPath)
	if err != nil {
		return nil, fmt.Errorf("provider creation failed: %w", err)
	}
	info := provider.GetDatabaseInfo()
	log.Info().
		Bool("city_db", info["city_db_loaded"].(bool)).
		Bool("asn_db", info["asn_db_loaded"].(bool)).
		Msg("GeoIP provider initialized successfully")
	return provider, nil
}

func main() {
	log.Info().Msg("Starting peerlogger")

	blacklist := crawler.NewBlacklist(appCfg.IPBlacklist, appCfg.PubkeyBlacklist)
	nodeRepo, err := initDB()
	if err != nil {
		log.Fatal().Err(err).Str("db_url", envCfg.DBURL).Msg("Database initialization failed")
	}
	geoIP, err := initGeoIP()
	if err != nil {
		log.Fatal().Err(err).Str("city_db", appCfg.GeoIPCityDBPath).Str("asn_db", appCfg.GeoIPASNDBPath).Msg("GeoIP provider initialization failed")
	}

	crawlerCfg := &crawler.Config{
		ConfigHash:        appCfgHash,
		NodeRepository:    nodeRepo,
		Blacklist:         blacklist,
		GeoIP:             geoIP,
		CrawlInterval:     appCfg.CrawlInterval,
		DiscoveryTimeout:  appCfg.DiscoveryTimeout,
		MaxParallelCrawls: appCfg.MaxParallelCrawls,
	}
	c := crawler.NewCrawler(crawlerCfg)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		c.Start(ctx)
	}()

	<-sigChan
	log.Info().Msg("Shutdown signal received, stopping crawler...")
	time.Sleep(5 * time.Second) // allow some time for graceful shutdown
	log.Info().Msg("Crawler stopped, goodbye")
}
