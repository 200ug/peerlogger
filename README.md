# peerlogger

Lightweight version of the Ethernets [node-crawler](https://github.com/ethereum/node-crawler):

- Similarly based on the [devp2p](https://github.com/ethereum/go-ethereum/tree/master/cmd/devp2p) tool
- Contains data processing logic compatible with PostgreSQL
- No special filtering features except for basic dynamically loaded blacklist
- Compatible with GeoIP databases, such as GeoLite2 from MaxMind
- Containerization with Podman
