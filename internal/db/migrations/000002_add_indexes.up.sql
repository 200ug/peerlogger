-- el nodes performance
CREATE INDEX IF NOT EXISTS idx_el_nodes_last_seen ON el_nodes (last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_el_nodes_first_seen ON el_nodes (first_seen DESC);
CREATE INDEX IF NOT EXISTS idx_el_nodes_ip_address ON el_nodes (ip_address);
CREATE INDEX IF NOT EXISTS idx_el_nodes_client_name ON el_nodes (client_name);
CREATE INDEX IF NOT EXISTS idx_el_nodes_protocol_version ON el_nodes (protocol_version);
CREATE INDEX IF NOT EXISTS idx_el_nodes_country_code ON el_nodes (country_code);
CREATE INDEX IF NOT EXISTS idx_el_nodes_connection_status ON el_nodes (connection_status);
CREATE INDEX IF NOT EXISTS idx_el_nodes_failure_count ON el_nodes (failure_count);
CREATE INDEX IF NOT EXISTS idx_el_nodes_network_id ON el_nodes (network_id);
CREATE INDEX IF NOT EXISTS idx_el_nodes_chain_id ON el_nodes (chain_id);

-- composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_el_nodes_client_version ON el_nodes (client_name, client_version);
CREATE INDEX IF NOT EXISTS idx_el_nodes_geo ON el_nodes (country_code, city_name);
CREATE INDEX IF NOT EXISTS idx_el_nodes_status_last_seen ON el_nodes (connection_status, last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_el_nodes_protocol_status ON el_nodes (protocol_version, connection_status);
CREATE INDEX IF NOT EXISTS idx_el_nodes_network_status ON el_nodes (network_id, connection_status);
