package crawler

import "net"

type GeoIP struct {
	// todo: add variables necessary for storing the geoip db handler data
	// ** this is a plceholder **
}

type GeoIPInfo struct {
	Country string
	City    string
	Lat     float64
	Long    float64
}

func NewGeoIP() *GeoIP {
	// todo: init geoip db (e.g. geolite2 from maxmind)
	// ** this is a plceholder **
	return &GeoIP{}
}

func (g *GeoIP) Lookup(ip net.IP) *GeoIPInfo {
	// todo: look up ip in geoip db (https://github.com/oschwald/geoip2-golang)
	// ** this is a plceholder **

	return &GeoIPInfo{
		Country: "Unknown",
		City:    "Unknown",
		Lat:     0.0,
		Long:    0.0,
	}
}
