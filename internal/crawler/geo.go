package crawler

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/oschwald/geoip2-golang/v2"
	"github.com/rs/zerolog/log"
)

type GeoIP struct {
	cityDB *geoip2.Reader
	asnDB  *geoip2.Reader
	mu     sync.RWMutex
}

type GeoData struct {
	CountryName *string
	CountryCode *string
	CityName    *string
	ASNumber    *int64
}

// cityDBPath: path to GeoLite2-City.mmdb
// asnDBPath: path to GeoLite2-ASN.mmdb
func NewGeoIP(cityDBPath, asnDBPath string) (*GeoIP, error) {
	provider := &GeoIP{}
	if cityDBPath != "" {
		if err := provider.LoadCityDatabase(cityDBPath); err != nil {
			return nil, fmt.Errorf("failed to load city database: %w", err)
		}
	}
	if asnDBPath != "" {
		if err := provider.LoadASNDatabase(asnDBPath); err != nil {
			return nil, fmt.Errorf("failed to load ASN database: %w", err)
		}
	}
	return provider, nil
}

func (g *GeoIP) LoadCityDatabase(dbPath string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// close existing db if open
	if g.cityDB != nil {
		if err := g.cityDB.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close existing city database")
		}
	}
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open city database at %s: %w", dbPath, err)
	}
	g.cityDB = db
	log.Info().Str("path", dbPath).Msg("GeoLite2 City database loaded successfully")

	return nil
}

func (g *GeoIP) LoadASNDatabase(dbPath string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// close existing db if open
	if g.asnDB != nil {
		if err := g.asnDB.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close existing ASN database")
		}
	}
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open ASN database at %s: %w", dbPath, err)
	}
	g.asnDB = db
	log.Info().Str("path", dbPath).Msg("GeoLite2 ASN database loaded successfully")

	return nil
}

func (g *GeoIP) Lookup(ip netip.Addr) (*GeoData, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	geoData := &GeoData{}
	if g.cityDB != nil {
		cityRecord, err := g.cityDB.City(ip)
		if err != nil {
			log.Debug().Err(err).Str("ip", ip.String()).Msg("City lookup failed")
		} else {
			g.extractCityData(cityRecord, geoData)
		}
	}
	if g.asnDB != nil {
		asnRecord, err := g.asnDB.ASN(ip)
		if err != nil {
			log.Debug().Err(err).Str("ip", ip.String()).Msg("ASN lookup failed")
		} else {
			g.extractASNData(asnRecord, geoData)
		}
	}
	return geoData, nil
}

func (g *GeoIP) extractCityData(record *geoip2.City, geoData *GeoData) {
	if len(record.Country.Names.English) > 0 {
		geoData.CountryName = &record.Country.Names.English
	}
	if len(record.Country.ISOCode) > 0 {
		geoData.CountryCode = &record.Country.ISOCode
	}
	if len(record.City.Names.English) > 0 {
		geoData.CityName = &record.City.Names.English
	}
}

func (g *GeoIP) extractASNData(record *geoip2.ASN, geoData *GeoData) {
	if record.AutonomousSystemNumber > 0 {
		asNum := int64(record.AutonomousSystemNumber)
		geoData.ASNumber = &asNum
	}
}

func (g *GeoIP) GetDatabaseInfo() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	info := map[string]interface{}{
		"city_db_loaded": g.cityDB != nil,
		"asn_db_loaded":  g.asnDB != nil,
	}

	return info
}

func (g *GeoIP) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.cityDB != nil {
		if err := g.cityDB.Close(); err != nil {
			return err
		}
		g.cityDB = nil
	}
	if g.asnDB != nil {
		if err := g.asnDB.Close(); err != nil {
			return err
		}
		g.asnDB = nil
	}

	return nil
}
