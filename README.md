# peerlogger

Lightweight version of the ethernets.io [node-crawler](https://github.com/ethereum/node-crawler):

- Crawls the Ethereum network for both discv4 & discv5 protocols with dial-only connections (crawler dialing to other nodes, but accepting no connections from other nodes)
- Focuses on gathering bulk data, i.e. doesn't contain advanced filtering options except for a simple dynamically loaded blacklist
- Contains data processing logic compatible with PostgreSQL
- Compatible with GeoIP databases, such as GeoLite2 from MaxMind
- Containerization with Podman

## Quick Start

1. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your database URL and GeoIP database paths
   ```

2. **Build the application:**
   ```bash
   go build .
   ```

3. **Run the crawler:**
   ```bash
   ./peerlogger
   ```

## Environment Variables

- `LOG_LEVEL`: Log level (debug, info, warn, error)
- `DB_URL`: PostgreSQL connection string
- `IP_BLACKLIST_PATH`: Optional path to IP blacklist JSON file
- `PUBKEY_BLACKLIST_PATH`: Optional path to pubkey blacklist JSON file  
- `GEOIP_CITY_DB_PATH`: Path to GeoLite2 City database
- `GEOIP_ASN_DB_PATH`: Path to GeoLite2 ASN database
- `ENR_DB_PATH`: Path for storing discovered node records

## Blacklist File Format

Blacklist files should be JSON with the following structure:

**IP Blacklist (`ip.json`):**
```json
{
  "ip_blacklists": [
    "127.0.0.1",
    "192.168.0.0/16", 
    "10.0.0.0/8"
  ]
}
```

**Pubkey Blacklist (`pubkey.json`):**
```json
{
  "pubkey_blacklists": [
    "0x1234567890abcdef...",
    "malicious_pubkey_1"
  ]
}
```

## Usage

1. Define app configuration and secrets in `config.json` and `.env` (see the template)
2. For GeoIP support, add `GeoLite2-City.mmdb` and `GeoLite2-ASN.mmdb` to `geoip/` ([MaxMind](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data/))
3. Make sure `podman-compose` is installed, then use `./launch.sh up` and `./launch.sh down` to start and stop the project's containers
