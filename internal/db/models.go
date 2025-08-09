package db

import (
	"database/sql"
	"time"
)

type NodeStatus string

const (
	NodeStatusUnknown    NodeStatus = "unknown"
	NodeStatusOnline     NodeStatus = "online"
	NodeStatusOffline    NodeStatus = "offline"
	NodeStatusConnecting NodeStatus = "connecting"
	NodeStatusError      NodeStatus = "error"
)

type Node struct {
	ID               int64          `db:"id"`
	NodeID           string         `db:"node_id"`           // enr node id or public key
	IPAddress        string         `db:"ip_address"`        // ip address
	TCPPort          int            `db:"tcp_port"`          // tcp port (eth protocol)
	UDPPort          int            `db:"udp_port"`          // udp port (discovery protocol)
	ClientName       string         `db:"client_name"`       // client name/version (e.g., "Geth/v1.10.26")
	ClientVersion    string         `db:"client_version"`    // detailed version info
	ProtocolVersion  int            `db:"protocol_version"`  // ethereum protocol version (66, 67, 68, etc.)
	ForkID           sql.NullString `db:"fork_id"`           // fork identifier (for network compatibility)
	HeadHash         sql.NullString `db:"head_hash"`         // current head block hash (sync status)
	NetworkID        sql.NullInt64  `db:"network_id"`        // network identifier
	ChainID          sql.NullInt64  `db:"chain_id"`          // chain identifier
	FirstSeen        time.Time      `db:"first_seen"`        // first time node was discovered
	LastSeen         time.Time      `db:"last_seen"`         // last time node was seen
	CountryCode      sql.NullString `db:"country_code"`      // iso country code (e.g. "us")
	CityName         sql.NullString `db:"city_name"`         // city name
	ASNumber         sql.NullInt64  `db:"as_number"`         // autonomous system number
	ConnectionStatus NodeStatus     `db:"connection_status"` // connection status
	FailureCount     int            `db:"failure_count"`     // number of consecutive failures
	PingRTT          sql.NullInt64  `db:"ping_rtt"`          // ping round-trip time in microseconds
	UpdatedAt        time.Time      `db:"updated_at"`        // last update timestamp
	CreatedAt        time.Time      `db:"created_at"`        // record creation timestamp
}

// Representation for upserting an execution layer (EL) node to the database.
type NodeUpsert struct {
	NodeID           string
	IPAddress        string
	TCPPort          int
	UDPPort          int
	ClientName       string
	ClientVersion    string
	ProtocolVersion  int
	ForkID           *string
	HeadHash         *string
	NetworkID        *int64
	ChainID          *int64
	LastSeen         time.Time
	FirstSeen        time.Time
	CountryCode      *string
	CityName         *string
	ASNumber         *int64
	ConnectionStatus NodeStatus
	FailureCount     int
	PingRTT          *int64
}

// Provider of db operations for execution layer (EL) nodes.
type NodeRepository struct {
	db *Database
}

func NewNodeRepository(db *Database) *NodeRepository {
	return &NodeRepository{db: db}
}

func (r *NodeRepository) UpsertNode(node NodeUpsert) error {
	query := `
		INSERT INTO el_nodes (
			node_id, ip_address, tcp_port, udp_port, client_name, client_version,
			protocol_version, fork_id, head_hash, network_id, chain_id,
			first_seen, last_seen, country_code, city_name, as_number,
			connection_status, failure_count, ping_rtt, updated_at, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, NOW(), NOW()
		)
		ON CONFLICT (node_id) DO UPDATE SET
			ip_address = EXCLUDED.ip_address,
			tcp_port = EXCLUDED.tcp_port,
			udp_port = EXCLUDED.udp_port,
			client_name = EXCLUDED.client_name,
			client_version = EXCLUDED.client_version,
			protocol_version = EXCLUDED.protocol_version,
			fork_id = EXCLUDED.fork_id,
			head_hash = EXCLUDED.head_hash,
			network_id = EXCLUDED.network_id,
			chain_id = EXCLUDED.chain_id,
			last_seen = EXCLUDED.last_seen,
			country_code = EXCLUDED.country_code,
			city_name = EXCLUDED.city_name,
			as_number = EXCLUDED.as_number,
			connection_status = EXCLUDED.connection_status,
			failure_count = CASE 
				WHEN EXCLUDED.connection_status = 'online' THEN 0
				ELSE el_nodes.failure_count + EXCLUDED.failure_count
			END,
			ping_rtt = EXCLUDED.ping_rtt,
			updated_at = NOW()
	`

	_, err := r.db.Exec(query,
		node.NodeID, node.IPAddress, node.TCPPort, node.UDPPort,
		node.ClientName, node.ClientVersion, node.ProtocolVersion,
		node.ForkID, node.HeadHash, node.NetworkID, node.ChainID,
		node.FirstSeen, node.LastSeen,
		node.CountryCode, node.CityName, node.ASNumber,
		node.ConnectionStatus, node.FailureCount, node.PingRTT,
	)

	return err
}

func (r *NodeRepository) GetNodeByNodeID(nodeID string) (*Node, error) {
	query := `
		SELECT id, node_id, ip_address, tcp_port, udp_port, client_name, client_version,
		       protocol_version, fork_id, head_hash, network_id, chain_id,
		       first_seen, last_seen, country_code, city_name, as_number,
		       connection_status, failure_count, ping_rtt, updated_at, created_at
		FROM el_nodes
		WHERE node_id = $1
	`

	var node Node
	err := r.db.QueryRow(query, nodeID).Scan(
		&node.ID, &node.NodeID, &node.IPAddress, &node.TCPPort, &node.UDPPort,
		&node.ClientName, &node.ClientVersion, &node.ProtocolVersion,
		&node.ForkID, &node.HeadHash, &node.NetworkID, &node.ChainID,
		&node.FirstSeen, &node.LastSeen, &node.CountryCode, &node.CityName,
		&node.ASNumber, &node.ConnectionStatus, &node.FailureCount,
		&node.PingRTT, &node.UpdatedAt, &node.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func (r *NodeRepository) GetAllNodes(limit, offset int) ([]*Node, error) {
	query := `
		SELECT id, node_id, ip_address, tcp_port, udp_port, client_name, client_version,
		       protocol_version, fork_id, head_hash, network_id, chain_id,
		       first_seen, last_seen, country_code, city_name, as_number,
		       connection_status, failure_count, ping_rtt, updated_at, created_at
		FROM el_nodes
		ORDER BY last_seen DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		err := rows.Scan(
			&node.ID, &node.NodeID, &node.IPAddress, &node.TCPPort, &node.UDPPort,
			&node.ClientName, &node.ClientVersion, &node.ProtocolVersion,
			&node.ForkID, &node.HeadHash, &node.NetworkID, &node.ChainID,
			&node.FirstSeen, &node.LastSeen, &node.CountryCode, &node.CityName,
			&node.ASNumber, &node.ConnectionStatus, &node.FailureCount,
			&node.PingRTT, &node.UpdatedAt, &node.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, rows.Err()
}

func (r *NodeRepository) GetActiveNodes(since time.Duration) ([]*Node, error) {
	query := `
		SELECT id, node_id, ip_address, tcp_port, udp_port, client_name, client_version,
		       protocol_version, fork_id, head_hash, network_id, chain_id,
		       first_seen, last_seen, country_code, city_name, as_number,
		       connection_status, failure_count, ping_rtt, updated_at, created_at
		FROM el_nodes
		WHERE last_seen >= NOW() - INTERVAL '$1 seconds'
		ORDER BY last_seen DESC
	`

	rows, err := r.db.Query(query, int64(since.Seconds()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		err := rows.Scan(
			&node.ID, &node.NodeID, &node.IPAddress, &node.TCPPort, &node.UDPPort,
			&node.ClientName, &node.ClientVersion, &node.ProtocolVersion,
			&node.ForkID, &node.HeadHash, &node.NetworkID, &node.ChainID,
			&node.FirstSeen, &node.LastSeen, &node.CountryCode, &node.CityName,
			&node.ASNumber, &node.ConnectionStatus, &node.FailureCount,
			&node.PingRTT, &node.UpdatedAt, &node.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, rows.Err()
}

func (r *NodeRepository) UpdateNodeStatus(nodeID string, status NodeStatus, failureCount int) error {
	query := `
		UPDATE el_nodes
		SET connection_status = $1,
		    failure_count = $2,
		    updated_at = NOW()
		WHERE node_id = $3
	`

	_, err := r.db.Exec(query, status, failureCount, nodeID)
	return err
}

func (r *NodeRepository) DeleteOldNodes(olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM el_nodes
		WHERE last_seen < NOW() - INTERVAL '$1 seconds'
	`

	result, err := r.db.Exec(query, int64(olderThan.Seconds()))
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *NodeRepository) GetNodesByClient(clientName string, limit, offset int) ([]*Node, error) {
	query := `
		SELECT id, node_id, ip_address, tcp_port, udp_port, client_name, client_version,
		       protocol_version, fork_id, head_hash, network_id, chain_id,
		       first_seen, last_seen, country_code, city_name, as_number,
		       connection_status, failure_count, ping_rtt, updated_at, created_at
		FROM el_nodes
		WHERE client_name ILIKE $1
		ORDER BY last_seen DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, "%"+clientName+"%", limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		err := rows.Scan(
			&node.ID, &node.NodeID, &node.IPAddress, &node.TCPPort, &node.UDPPort,
			&node.ClientName, &node.ClientVersion, &node.ProtocolVersion,
			&node.ForkID, &node.HeadHash, &node.NetworkID, &node.ChainID,
			&node.FirstSeen, &node.LastSeen, &node.CountryCode, &node.CityName,
			&node.ASNumber, &node.ConnectionStatus, &node.FailureCount,
			&node.PingRTT, &node.UpdatedAt, &node.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, rows.Err()
}

func (r *NodeRepository) GetNodesByCountry(countryCode string, limit, offset int) ([]*Node, error) {
	query := `
		SELECT id, node_id, ip_address, tcp_port, udp_port, client_name, client_version,
		       protocol_version, fork_id, head_hash, network_id, chain_id,
		       first_seen, last_seen, country_code, city_name, as_number,
		       connection_status, failure_count, ping_rtt, updated_at, created_at
		FROM el_nodes
		WHERE country_code = $1
		ORDER BY last_seen DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, countryCode, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		err := rows.Scan(
			&node.ID, &node.NodeID, &node.IPAddress, &node.TCPPort, &node.UDPPort,
			&node.ClientName, &node.ClientVersion, &node.ProtocolVersion,
			&node.ForkID, &node.HeadHash, &node.NetworkID, &node.ChainID,
			&node.FirstSeen, &node.LastSeen, &node.CountryCode, &node.CityName,
			&node.ASNumber, &node.ConnectionStatus, &node.FailureCount,
			&node.PingRTT, &node.UpdatedAt, &node.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, rows.Err()
}

func (r *NodeRepository) GetNodeStats() (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_nodes,
			COUNT(CASE WHEN connection_status = 'online' THEN 1 END) as online_nodes,
			COUNT(CASE WHEN connection_status = 'offline' THEN 1 END) as offline_nodes,
			COUNT(DISTINCT client_name) as unique_clients,
			COUNT(DISTINCT country_code) as unique_countries
		FROM el_nodes
	`

	var total, online, offline, clients, countries int64
	err := r.db.QueryRow(query).Scan(&total, &online, &offline, &clients, &countries)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_nodes":      total,
		"online_nodes":     online,
		"offline_nodes":    offline,
		"unique_clients":   clients,
		"unique_countries": countries,
	}

	return stats, nil
}

func (r *NodeRepository) GetNodesByProtocolVersion(version int, limit, offset int) ([]*Node, error) {
	query := `
		SELECT id, node_id, ip_address, tcp_port, udp_port, client_name, client_version,
		       protocol_version, fork_id, head_hash, network_id, chain_id,
		       first_seen, last_seen, country_code, city_name, as_number,
		       connection_status, failure_count, ping_rtt, updated_at, created_at
		FROM el_nodes
		WHERE protocol_version = $1
		ORDER BY last_seen DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, version, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		err := rows.Scan(
			&node.ID, &node.NodeID, &node.IPAddress, &node.TCPPort, &node.UDPPort,
			&node.ClientName, &node.ClientVersion, &node.ProtocolVersion,
			&node.ForkID, &node.HeadHash, &node.NetworkID, &node.ChainID,
			&node.FirstSeen, &node.LastSeen, &node.CountryCode, &node.CityName,
			&node.ASNumber, &node.ConnectionStatus, &node.FailureCount,
			&node.PingRTT, &node.UpdatedAt, &node.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, rows.Err()
}

func (r *NodeRepository) GetNodesByNetworkID(networkID int64, limit, offset int) ([]*Node, error) {
	query := `
		SELECT id, node_id, ip_address, tcp_port, udp_port, client_name, client_version,
		       protocol_version, fork_id, head_hash, network_id, chain_id,
		       first_seen, last_seen, country_code, city_name, as_number,
		       connection_status, failure_count, ping_rtt, updated_at, created_at
		FROM el_nodes
		WHERE network_id = $1
		ORDER BY last_seen DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, networkID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*Node
	for rows.Next() {
		var node Node
		err := rows.Scan(
			&node.ID, &node.NodeID, &node.IPAddress, &node.TCPPort, &node.UDPPort,
			&node.ClientName, &node.ClientVersion, &node.ProtocolVersion,
			&node.ForkID, &node.HeadHash, &node.NetworkID, &node.ChainID,
			&node.FirstSeen, &node.LastSeen, &node.CountryCode, &node.CityName,
			&node.ASNumber, &node.ConnectionStatus, &node.FailureCount,
			&node.PingRTT, &node.UpdatedAt, &node.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &node)
	}

	return nodes, rows.Err()
}

func (r *NodeRepository) GetOnlineNodesCount() (int64, error) {
	query := `SELECT COUNT(*) FROM el_nodes WHERE connection_status = 'online'`
	var count int64
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

func (r *NodeRepository) GetClientDistribution() (map[string]int64, error) {
	query := `
		SELECT client_name, COUNT(*) as count
		FROM el_nodes
		WHERE client_name IS NOT NULL
		GROUP BY client_name
		ORDER BY count DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	distribution := make(map[string]int64)
	for rows.Next() {
		var clientName string
		var count int64
		err := rows.Scan(&clientName, &count)
		if err != nil {
			return nil, err
		}
		distribution[clientName] = count
	}

	return distribution, rows.Err()
}

func (r *NodeRepository) GetCountryDistribution() (map[string]int64, error) {
	query := `
		SELECT country_code, COUNT(*) as count
		FROM el_nodes
		WHERE country_code IS NOT NULL
		GROUP BY country_code
		ORDER BY count DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	distribution := make(map[string]int64)
	for rows.Next() {
		var countryCode string
		var count int64
		err := rows.Scan(&countryCode, &count)
		if err != nil {
			return nil, err
		}
		distribution[countryCode] = count
	}

	return distribution, rows.Err()
}
