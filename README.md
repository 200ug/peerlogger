# peerlogger

Ethereum network peer discovery crawler

## Features

- Crawls Ethereum network using discv4 & discv5 protocols
- Stores discovered node data in PostgreSQL database
- Extracts client information (Geth, Besu, Erigon, etc.)
- GeoIP enrichment with country, city, and ASN data
- JSON-based IP and pubkey blacklisting
- Structured logging with configurable levels
- Graceful shutdown handling

## Usage

```bash
# Setup environment
cp .env.example .env
# Edit .env with your database URL and paths

# Use the included shellscript to spawn containers for crawler + PostgreSQL
chmod +x launch.sh
./launch.sh up    # Startup
./launch.sh down  # Shutdown
```
