CREATE TABLE IF NOT EXISTS el_nodes (
    id BIGSERIAL PRIMARY KEY,
    node_id VARCHAR(128) UNIQUE NOT NULL,
    ip_address INET NOT NULL,
    tcp_port INTEGER,
    udp_port INTEGER,
    client_name VARCHAR(128),
    client_version VARCHAR(128),
    protocol_version INTEGER,                -- ethereum protocol version (66, 67, 68, etc.)
    fork_id VARCHAR(64),                     -- fork identifier for network compatibility
    head_hash VARCHAR(66),                   -- current head block hash (0x prefixed)
    network_id BIGINT,                       -- network identifier (1=mainnet, 5=goerli, etc.)
    chain_id BIGINT,                         -- chain identifier
    first_seen TIMESTAMP WITH TIME ZONE NOT NULL,
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL,
    country_code CHAR(2),
    city_name VARCHAR(128),
    as_number BIGINT,
    connection_status VARCHAR(20) DEFAULT 'unknown',
    failure_count INTEGER DEFAULT 0,
    ping_rtt BIGINT,                         -- ping round-trip time in microseconds
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);