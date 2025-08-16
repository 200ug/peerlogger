# peerlogger

Ethereum network peer discovery crawler

## Features

- Crawls Ethereum network using discv4 & discv5 protocols
- Data processing logic for PostgreSQL
- Client information extraction
- GeoIP support with country, city, and ASN data
- Simple IP and pubkey blacklisting

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
