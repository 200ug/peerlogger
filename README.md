# peerlogger

Lightweight version of the Ethernets [node-crawler](https://github.com/ethereum/node-crawler):

- Similarly based on the [devp2p](https://github.com/ethereum/go-ethereum/tree/master/cmd/devp2p) tool
- Contains data processing logic compatible with PostgreSQL
- Focuses on gathering bulk data, i.e. doesn't contain advanced filtering options except for a simple dynamically loaded blacklist
- Compatible with GeoIP databases, such as GeoLite2 from MaxMind
- Containerization with Podman

## Usage

1. Define app configuration and secrets in `config.json` and `.env` (see the template)
2. For GeoIP support, add `GeoLite2-City.mmdb` and `GeoLite2-ASN.mmdb` to `geoip/` ([MaxMind](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data/))
3. Make sure `podman-compose` is installed, then use `./launch.sh up` and `./launch.sh down` to start and stop the project's containers
