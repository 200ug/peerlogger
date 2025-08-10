package crawler_test

import (
	"net/netip"
	"testing"

	"github.com/200ug/peerlogger/internal/crawler"
)

func TestNewGeoIP(t *testing.T) {
	// empty paths should not panic
	geo, err := crawler.NewGeoIP("", "")
	if err != nil {
		t.Fatalf("NewGeoIP with empty paths should not error: %v", err)
	}
	if geo == nil {
		t.Fatal("NewGeoIP should return non-nil GeoIP")
	}
	// db info when no dbs loaded
	info := geo.GetDatabaseInfo()
	if info["city_db_loaded"].(bool) != false {
		t.Error("City DB should not be loaded")
	}
	if info["asn_db_loaded"].(bool) != false {
		t.Error("ASN DB should not be loaded")
	}
}

func TestGeoIP_Close(t *testing.T) {
	geo, err := crawler.NewGeoIP("", "")
	if err != nil {
		t.Fatalf("Failed to create GeoIP: %v", err)
	}
	// close should not error even when no dbs loaded
	err = geo.Close()
	if err != nil {
		t.Errorf("Close should not error when no databases loaded: %v", err)
	}
	// close again should also not error
	err = geo.Close()
	if err != nil {
		t.Errorf("Second close should not error: %v", err)
	}
}

func TestGeoIP_GetDatabaseInfo(t *testing.T) {
	geo, err := crawler.NewGeoIP("", "")
	if err != nil {
		t.Fatalf("Failed to create GeoIP: %v", err)
	}
	info := geo.GetDatabaseInfo()
	if _, ok := info["city_db_loaded"]; !ok {
		t.Error("info should contain city_db_loaded")
	}
	if _, ok := info["asn_db_loaded"]; !ok {
		t.Error("info should contain asn_db_loaded")
	}
}

func TestGeoData_Structure(t *testing.T) {
	geoData := &crawler.GeoData{}
	// fields should be nil initially
	if geoData.CountryName != nil {
		t.Error("CountryName should be nil initially")
	}
	if geoData.CountryCode != nil {
		t.Error("CountryCode should be nil initially")
	}
	if geoData.CityName != nil {
		t.Error("CityName should be nil initially")
	}
	if geoData.ASNumber != nil {
		t.Error("ASNumber should be nil initially")
	}
	country := "United States"
	countryCode := "US"
	city := "New York"
	asNum := int64(15169)
	geoData = &crawler.GeoData{
		CountryName: &country,
		CountryCode: &countryCode,
		CityName:    &city,
		ASNumber:    &asNum,
	}
	if geoData.CountryName == nil || *geoData.CountryName != country {
		t.Error("CountryName not set correctly")
	}
	if geoData.CountryCode == nil || *geoData.CountryCode != countryCode {
		t.Error("CountryCode not set correctly")
	}
	if geoData.CityName == nil || *geoData.CityName != city {
		t.Error("CityName not set correctly")
	}
	if geoData.ASNumber == nil || *geoData.ASNumber != asNum {
		t.Error("ASNumber not set correctly")
	}
}

func TestGeoIP_LookupWithNoDatabases(t *testing.T) {
	geo, err := crawler.NewGeoIP("", "")
	if err != nil {
		t.Fatalf("Failed to create GeoIP: %v", err)
	}
	// lookup with no dbs loaded
	ip, err := netip.ParseAddr("8.8.8.8")
	if err != nil {
		t.Fatalf("Failed to parse IP: %v", err)
	}
	result, err := geo.Lookup(ip)
	if err != nil {
		t.Errorf("Lookup should not error when no databases loaded: %v", err)
	}
	if result == nil {
		t.Error("Lookup should return non-nil result even when no databases loaded")
	}
	//lint:ignore SA5011
	if result.CountryName != nil {
		t.Error("CountryName should be nil when no databases loaded")
	}
	if result.CountryCode != nil {
		t.Error("CountryCode should be nil when no databases loaded")
	}
	if result.CityName != nil {
		t.Error("CityName should be nil when no databases loaded")
	}
	if result.ASNumber != nil {
		t.Error("ASNumber should be nil when no databases loaded")
	}
}

func TestGeoIP_ValidIPAddresses(t *testing.T) {
	tests := []struct {
		name  string
		ip    string
		valid bool
	}{
		{"valid ipv4", "8.8.8.8", true},
		{"valid ipv6", "2001:4860:4860::8888", true},
		{"invalid ip", "999.999.999.999", false},
		{"empty string", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := netip.ParseAddr(tt.ip)
			if tt.valid {
				if err != nil {
					t.Errorf("Expected valid IP %s, but got error: %v", tt.ip, err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected invalid IP %s, but parsing succeeded", tt.ip)
				}
				return // no need to test lookup for invalid ips
			}
			// lookup with valid ip (no dbs loaded)
			geo, err := crawler.NewGeoIP("", "")
			if err != nil {
				t.Fatalf("Failed to create GeoIP: %v", err)
			}
			result, err := geo.Lookup(ip)
			if err != nil {
				t.Errorf("Lookup should not error: %v", err)
			}
			if result == nil {
				t.Error("Lookup should return non-nil result")
			}
		})
	}
}

func TestGeoIP_LoadDatabaseErrors(t *testing.T) {
	geo := &crawler.GeoIP{}
	// loading non-existent db files
	err := geo.LoadCityDatabase("/non/existent/path/GeoLite2-City.mmdb")
	if err == nil {
		t.Error("LoadCityDatabase should error for non-existent file")
	}
	err = geo.LoadASNDatabase("/non/existent/path/GeoLite2-ASN.mmdb")
	if err == nil {
		t.Error("LoadASNDatabase should error for non-existent file")
	}
}

// Integration test helper (**requires actual database files**)
func TestGeoIP_WithMockData(t *testing.T) {
	geo, err := crawler.NewGeoIP("../../geoip/GeoLite2-City.mmdb", "../../geoip/GeoLite2-ASN.mmdb")
	if err != nil {
		t.Skipf("Skipping test - GeoIP databases not available: %v", err)
	}
	defer geo.Close()
	info := geo.GetDatabaseInfo()
	if !info["city_db_loaded"].(bool) || !info["asn_db_loaded"].(bool) {
		t.Skip("Skipping Cloudflare test - databases not loaded")
	}
	testIPs := []string{
		"69.251.59.46",   // nodes/bcd62bfba5f0edf981429ccd8c735c857940def1bd61787995e5d956f8a96e1d
		"1.1.1.1",        // cloudflare dns
		"140.233.190.22", // internet magnate
		"208.67.222.222", // opendns
	}

	for _, ipStr := range testIPs {
		t.Run(ipStr, func(t *testing.T) {
			ip, err := netip.ParseAddr(ipStr)
			if err != nil {
				t.Fatalf("Failed to parse IP %s: %v", ipStr, err)
			}
			result, err := geo.Lookup(ip)
			if err != nil {
				t.Fatalf("Lookup failed for %s: %v", ipStr, err)
			}
			if result == nil {
				t.Fatalf("Result should not be nil for %s", ipStr)
			}
			t.Logf("Results for %s:", ipStr)

			if result.CountryName != nil {
				if len(*result.CountryName) == 0 {
					t.Errorf("CountryName should not be empty string for %s", ipStr)
				} else {
					t.Logf("Country: %s", *result.CountryName)
				}
			} else {
				t.Log("Country: <no data>")
			}

			if result.CountryCode != nil {
				if len(*result.CountryCode) == 0 {
					t.Errorf("CountryCode should not be empty string for %s", ipStr)
				} else {
					t.Logf("Country Code: %s", *result.CountryCode)
				}
			} else {
				t.Log("Country Code: <no data>")
			}

			if result.CityName != nil {
				t.Logf("City: %s", *result.CityName)
			} else {
				t.Log("City: <no data>")
			}

			if result.ASNumber != nil {
				if *result.ASNumber <= 0 {
					t.Errorf("ASNumber should be positive for %s, got %d", ipStr, *result.ASNumber)
				} else {
					t.Logf("AS Number: %d", *result.ASNumber)
				}
			} else {
				t.Log("AS Number: <no data>")
			}
		})
	}
}
