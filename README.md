# peerlogger

Lightweight version of the Ethernets [node-crawler](https://github.com/ethereum/node-crawler):

- Similarly based on the [devp2p](https://github.com/ethereum/go-ethereum/tree/master/cmd/devp2p) tool
- Contains data processing logic compatible with PostgreSQL
- Focuses on gathering bulk data, i.e. doesn't contain advanced filtering options except for a simple dynamically loaded blacklist
- Compatible with GeoIP databases, such as GeoLite2 from MaxMind
- Containerization with Podman

## Usage

The database password (`DB_PASSWORD`) must be specified in `.env`. It's also highly recommended to specify the `ip_blacklist` and `pubkey_blacklist` sections (along with other configurable variables) to the `config.json` file.

After this the project's containers can be managed with the included `launch.sh` shellscript (`./launch.sh up`/`./launch.sh down`).
