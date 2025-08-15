package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/200ug/peerlogger/internal/common"
	"github.com/200ug/peerlogger/internal/crawler"
	"github.com/200ug/peerlogger/internal/db"
	"github.com/200ug/peerlogger/internal/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

const (
	peerloggerVersion = "v0.1.1"
)

var (
	config     *util.EnvConfig
	appCfgHash [32]byte
)

func init() {
	config = util.LoadEnv()
	logLevel, _ := zerolog.ParseLevel(config.LogLevel) // should already be verified
	zerolog.SetGlobalLevel(logLevel)
	log.Debug().Msg("Primary initialization done")
}

func initDB() (*sql.DB, error) {
	database, err := sql.Open("postgres", config.DBURL)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// Test the connection
	if err := database.Ping(); err != nil {
		database.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Create the database schema if needed
	if err := db.CreateDB(database); err != nil {
		log.Warn().Err(err).Msg("Database schema creation failed (this may be expected if schema already exists)")
	}

	log.Info().Msg("Database connection established successfully")
	return database, nil
}

func initGeoIP() (*util.GeoIP, error) {
	if config.GeoIPCityDBPath == "" && config.GeoIPASNDBPath == "" {
		log.Info().Msg("No GeoIP database paths configured, skipping GeoIP initialization")
		return nil, nil
	}

	provider, err := util.NewGeoIP(config.GeoIPCityDBPath, config.GeoIPASNDBPath)
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

func initBlacklist() *util.Blacklist {
	// load blacklists if the paths are defined in .env
	var ipBlacklist []string
	if config.IPBlacklistPath != "" {
		ipBlacklist = util.LoadJSONList(config.IPBlacklistPath, "ip_blacklists")
	}
	var pubkeyBlacklist []string
	if config.PubkeyBlacklistPath != "" {
		pubkeyBlacklist = util.LoadJSONList(config.PubkeyBlacklistPath, "pubkey_blacklists")
	}

	blacklist := util.NewBlacklist(ipBlacklist, pubkeyBlacklist)
	ips, ipNets, pubkeys := blacklist.GetStats()
	log.Info().
		Int("ips", ips).
		Int("ipNets", ipNets).
		Int("pubkeys", pubkeys).
		Str("ip_file", config.IPBlacklistPath).
		Str("pubkey_file", config.PubkeyBlacklistPath).
		Msg("Blacklist initialized")

	return blacklist
}

func printStartupInfo() {
	log.Info().
		Str("version", peerloggerVersion).
		Int("gomaxprocs", runtime.GOMAXPROCS(0)).
		Int("num_cpu", runtime.NumCPU()).
		Str("go_version", runtime.Version()).
		Msg("Starting peerlogger")
}

func setupSignalHandling(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Info().Str("signal", sig.String()).Msg("Shutdown signal received")
				cancel()
				return
			}
		}
	}()
}

func main() {
	printStartupInfo()

	// Initialize blacklist
	_ = initBlacklist() // blacklist initialized but not used in this demo
	log.Info().Msg("Blacklist initialization completed")

	// Initialize database
	database, err := initDB()
	if err != nil {
		log.Fatal().Err(err).Str("db_url", config.DBURL).Msg("Database initialization failed")
	}
	defer database.Close()

	// Initialize GeoIP provider
	geoIP, err := initGeoIP()
	if err != nil {
		log.Fatal().Err(err).
			Str("city_db", config.GeoIPCityDBPath).
			Str("asn_db", config.GeoIPASNDBPath).
			Msg("GeoIP provider initialization failed")
	}
	if geoIP != nil {
		defer geoIP.Close()
	}

	// Initialize crawler components
	c := &crawler.Crawler{
		NetworkID:  1, // Ethereum mainnet
		NodeURL:    "", // Can be set from config if needed
		ListenAddr: ":30303",
		NodeKey:    "",
		Bootnodes:  []string{},
		Timeout:    30 * time.Second,
		Workers:    16,
		Sepolia:    false,
		Hoodi:      false,
	}

	log.Info().Msg("Crawler initialized successfully")

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupSignalHandling(cancel)

	// Demo crawling functionality
	log.Info().Msg("Starting peer crawler demo...")
	
	// Create an empty initial node set
	inputSet := make(common.NodeSet)
	
	// Run a crawl round (this is a simplified demo)
	go func() {
		ticker := time.NewTicker(60 * time.Second) // Crawl every minute
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				log.Info().Msg("Running crawl round...")
				
				// Run the crawler
				results := c.CrawlRound(inputSet, database, geoIP)
				
				log.Info().
					Int("discovered_nodes", len(results)).
					Msg("Crawl round completed")
					
				// Update inputSet with discovered nodes for next round
				inputSet = results
				
			case <-ctx.Done():
				log.Info().Msg("Crawler stopping...")
				return
			}
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Info().Msg("Shutdown signal received, stopping crawler...")
	
	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)
	log.Info().Msg("Peerlogger stopped")
}
