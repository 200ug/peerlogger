package db_test

import (
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"github.com/200ug/peerlogger/internal/db"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestNodeStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   db.NodeStatus
		expected string
	}{
		{"unknown status", db.NodeStatusUnknown, "unknown"},
		{"online status", db.NodeStatusOnline, "online"},
		{"offline status", db.NodeStatusOffline, "offline"},
		{"connecting status", db.NodeStatusConnecting, "connecting"},
		{"error status", db.NodeStatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestNewNodeRepository(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	if repo == nil {
		t.Error("Expected non-nil NodeRepository")
	}
}

func TestNodeRepository_UpsertNode(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	tests := []struct {
		name    string
		node    db.NodeUpsert
		wantErr bool
	}{
		{
			name: "successful upsert",
			node: db.NodeUpsert{
				NodeID:           "test-node-id",
				IPAddress:        "192.168.1.1",
				TCPPort:          30303,
				UDPPort:          30303,
				ClientName:       "Geth",
				ClientVersion:    "1.10.26",
				ProtocolVersion:  67,
				ForkID:           stringPtr("0x123"),
				HeadHash:         stringPtr("0xabc"),
				NetworkID:        int64Ptr(1),
				ChainID:          int64Ptr(1),
				LastSeen:         time.Now(),
				FirstSeen:        time.Now(),
				CountryCode:      stringPtr("US"),
				CityName:         stringPtr("New York"),
				ASNumber:         int64Ptr(15169),
				ConnectionStatus: db.NodeStatusOnline,
				FailureCount:     0,
				PingRTT:          int64Ptr(50000),
			},
			wantErr: false,
		},
		{
			name: "upsert with nil values",
			node: db.NodeUpsert{
				NodeID:           "test-node-id-2",
				IPAddress:        "10.0.0.1",
				TCPPort:          30303,
				UDPPort:          30303,
				ClientName:       "Geth",
				ClientVersion:    "1.10.26",
				ProtocolVersion:  67,
				LastSeen:         time.Now(),
				FirstSeen:        time.Now(),
				ConnectionStatus: db.NodeStatusOnline,
				FailureCount:     0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				mock.ExpectExec("INSERT INTO el_nodes").
					WithArgs(
						tt.node.NodeID, tt.node.IPAddress, tt.node.TCPPort, tt.node.UDPPort,
						tt.node.ClientName, tt.node.ClientVersion, tt.node.ProtocolVersion,
						tt.node.ForkID, tt.node.HeadHash, tt.node.NetworkID, tt.node.ChainID,
						sqlmock.AnyArg(), tt.node.LastSeen,
						tt.node.CountryCode, tt.node.CityName, tt.node.ASNumber,
						tt.node.ConnectionStatus, tt.node.FailureCount, tt.node.PingRTT,
					).WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.UpsertNode(tt.node)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestNodeRepository_GetNodeByNodeID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	tests := []struct {
		name     string
		nodeID   string
		mockRows *sqlmock.Rows
		wantErr  bool
		errType  error
	}{
		{
			name:   "successful get",
			nodeID: "test-node-id",
			mockRows: sqlmock.NewRows([]string{
				"id", "node_id", "ip_address", "tcp_port", "udp_port", "client_name", "client_version",
				"protocol_version", "fork_id", "head_hash", "network_id", "chain_id",
				"first_seen", "last_seen", "country_code", "city_name", "as_number",
				"connection_status", "failure_count", "ping_rtt", "updated_at", "created_at",
			}).AddRow(
				1, "test-node-id", "192.168.1.1", 30303, 30303, "Geth", "1.10.26",
				67, "0x123", "0xabc", 1, 1,
				time.Now(), time.Now(), "US", "New York", 15169,
				"online", 0, 50000, time.Now(), time.Now(),
			),
			wantErr: false,
		},
		{
			name:     "node not found",
			nodeID:   "nonexistent-node",
			mockRows: nil,
			wantErr:  true,
			errType:  sql.ErrNoRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := mock.ExpectQuery("SELECT (.+) FROM el_nodes WHERE node_id = \\$1")
			if tt.mockRows != nil {
				query.WithArgs(tt.nodeID).WillReturnRows(tt.mockRows)
			} else {
				query.WithArgs(tt.nodeID).WillReturnError(tt.errType)
			}

			node, err := repo.GetNodeByNodeID(tt.nodeID)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if node != nil {
					t.Error("Expected nil node on error")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if node == nil {
					t.Error("Expected non-nil node")
				} else {
					if node.NodeID != tt.nodeID {
						t.Errorf("Expected NodeID %s, got %s", tt.nodeID, node.NodeID)
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestNodeRepository_GetAllNodes(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	tests := []struct {
		name     string
		limit    int
		offset   int
		mockRows *sqlmock.Rows
		wantErr  bool
		expected int
	}{
		{
			name:   "successful get with results",
			limit:  10,
			offset: 0,
			mockRows: sqlmock.NewRows([]string{
				"id", "node_id", "ip_address", "tcp_port", "udp_port", "client_name", "client_version",
				"protocol_version", "fork_id", "head_hash", "network_id", "chain_id",
				"first_seen", "last_seen", "country_code", "city_name", "as_number",
				"connection_status", "failure_count", "ping_rtt", "updated_at", "created_at",
			}).AddRow(
				1, "node-1", "192.168.1.1", 30303, 30303, "Geth", "1.10.26",
				67, "0x123", "0xabc", 1, 1,
				time.Now(), time.Now(), "US", "New York", 15169,
				"online", 0, 50000, time.Now(), time.Now(),
			).AddRow(
				2, "node-2", "192.168.1.2", 30303, 30303, "Nethermind", "1.14.0",
				68, "0x456", "0xdef", 1, 1,
				time.Now(), time.Now(), "CA", "Toronto", 577,
				"offline", 1, 75000, time.Now(), time.Now(),
			),
			wantErr:  false,
			expected: 2,
		},
		{
			name:   "empty results",
			limit:  10,
			offset: 0,
			mockRows: sqlmock.NewRows([]string{
				"id", "node_id", "ip_address", "tcp_port", "udp_port", "client_name", "client_version",
				"protocol_version", "fork_id", "head_hash", "network_id", "chain_id",
				"first_seen", "last_seen", "country_code", "city_name", "as_number",
				"connection_status", "failure_count", "ping_rtt", "updated_at", "created_at",
			}),
			wantErr:  false,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery("SELECT (.+) FROM el_nodes ORDER BY last_seen DESC LIMIT \\$1 OFFSET \\$2").
				WithArgs(tt.limit, tt.offset).
				WillReturnRows(tt.mockRows)

			nodes, err := repo.GetAllNodes(tt.limit, tt.offset)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(nodes) != tt.expected {
					t.Errorf("Expected %d nodes, got %d", tt.expected, len(nodes))
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestNodeRepository_UpdateNodeStatus(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	tests := []struct {
		name         string
		nodeID       string
		status       db.NodeStatus
		failureCount int
		wantErr      bool
	}{
		{
			name:         "successful update",
			nodeID:       "test-node-id",
			status:       db.NodeStatusOnline,
			failureCount: 0,
			wantErr:      false,
		},
		{
			name:         "update to offline",
			nodeID:       "test-node-id",
			status:       db.NodeStatusOffline,
			failureCount: 3,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectExec("UPDATE el_nodes SET connection_status = \\$1, failure_count = \\$2, updated_at = NOW\\(\\) WHERE node_id = \\$3").
				WithArgs(tt.status, tt.failureCount, tt.nodeID).
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := repo.UpdateNodeStatus(tt.nodeID, tt.status, tt.failureCount)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestNodeRepository_DeleteOldNodes(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	tests := []struct {
		name          string
		olderThan     time.Duration
		mockResult    driver.Result
		wantErr       bool
		expectedCount int64
	}{
		{
			name:          "successful deletion",
			olderThan:     24 * time.Hour,
			mockResult:    sqlmock.NewResult(0, 5),
			wantErr:       false,
			expectedCount: 5,
		},
		{
			name:          "no nodes to delete",
			olderThan:     1 * time.Hour,
			mockResult:    sqlmock.NewResult(0, 0),
			wantErr:       false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectExec("DELETE FROM el_nodes WHERE last_seen < NOW\\(\\) - INTERVAL '\\$1 seconds'").
				WithArgs(int64(tt.olderThan.Seconds())).
				WillReturnResult(tt.mockResult)

			count, err := repo.DeleteOldNodes(tt.olderThan)

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if count != tt.expectedCount {
				t.Errorf("Expected %d deleted rows, got %d", tt.expectedCount, count)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestNodeRepository_GetNodeStats(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	tests := []struct {
		name        string
		mockRow     *sqlmock.Rows
		wantErr     bool
		expectedMap map[string]interface{}
	}{
		{
			name: "successful stats",
			mockRow: sqlmock.NewRows([]string{"total_nodes", "online_nodes", "offline_nodes", "unique_clients", "unique_countries"}).
				AddRow(100, 85, 15, 5, 25),
			wantErr: false,
			expectedMap: map[string]interface{}{
				"total_nodes":      int64(100),
				"online_nodes":     int64(85),
				"offline_nodes":    int64(15),
				"unique_clients":   int64(5),
				"unique_countries": int64(25),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery("SELECT COUNT\\(\\*\\) as total_nodes").
				WillReturnRows(tt.mockRow)

			stats, err := repo.GetNodeStats()

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.wantErr {
				for key, expectedValue := range tt.expectedMap {
					if actualValue, exists := stats[key]; !exists {
						t.Errorf("Expected key %s not found in stats", key)
					} else if actualValue != expectedValue {
						t.Errorf("Expected %s = %v, got %v", key, expectedValue, actualValue)
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestNodeRepository_GetOnlineNodesCount(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	tests := []struct {
		name          string
		mockResult    *sqlmock.Rows
		wantErr       bool
		expectedCount int64
	}{
		{
			name:          "successful count",
			mockResult:    sqlmock.NewRows([]string{"count"}).AddRow(42),
			wantErr:       false,
			expectedCount: 42,
		},
		{
			name:          "zero count",
			mockResult:    sqlmock.NewRows([]string{"count"}).AddRow(0),
			wantErr:       false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM el_nodes WHERE connection_status = 'online'").
				WillReturnRows(tt.mockResult)

			count, err := repo.GetOnlineNodesCount()

			if tt.wantErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if count != tt.expectedCount {
				t.Errorf("Expected count %d, got %d", tt.expectedCount, count)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestNodeRepository_ErrorHandling(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	t.Run("UpsertNode database error", func(t *testing.T) {
		node := db.NodeUpsert{
			NodeID:           "test-node-id",
			IPAddress:        "192.168.1.1",
			TCPPort:          30303,
			UDPPort:          30303,
			ClientName:       "Geth",
			ClientVersion:    "1.10.26",
			ProtocolVersion:  67,
			LastSeen:         time.Now(),
			FirstSeen:        time.Now(),
			ConnectionStatus: db.NodeStatusOnline,
			FailureCount:     0,
		}

		mock.ExpectExec("INSERT INTO el_nodes").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		err := repo.UpsertNode(node)
		if err == nil {
			t.Error("Expected database error, got nil")
		}
	})

	t.Run("GetActiveNodes database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM el_nodes WHERE last_seen >= NOW\\(\\) - INTERVAL").
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(sql.ErrConnDone)

		nodes, err := repo.GetActiveNodes(24 * time.Hour)
		if err == nil {
			t.Error("Expected database error, got nil")
		}
		if nodes != nil {
			t.Error("Expected nil nodes on error")
		}
	})

	t.Run("GetClientDistribution database error", func(t *testing.T) {
		mock.ExpectQuery("SELECT client_name, COUNT\\(\\*\\) as count FROM el_nodes").
			WillReturnError(sql.ErrConnDone)

		distribution, err := repo.GetClientDistribution()
		if err == nil {
			t.Error("Expected database error, got nil")
		}
		if distribution != nil {
			t.Error("Expected nil distribution on error")
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestNodeRepository_EdgeCases(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	t.Run("GetAllNodes with zero limit", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM el_nodes ORDER BY last_seen DESC LIMIT \\$1 OFFSET \\$2").
			WithArgs(0, 0).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "node_id", "ip_address", "tcp_port", "udp_port", "client_name", "client_version",
				"protocol_version", "fork_id", "head_hash", "network_id", "chain_id",
				"first_seen", "last_seen", "country_code", "city_name", "as_number",
				"connection_status", "failure_count", "ping_rtt", "updated_at", "created_at",
			}))

		nodes, err := repo.GetAllNodes(0, 0)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(nodes) != 0 {
			t.Errorf("Expected 0 nodes, got %d", len(nodes))
		}
	})

	t.Run("GetActiveNodes with zero duration", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM el_nodes WHERE last_seen >= NOW\\(\\) - INTERVAL").
			WithArgs(int64(0)).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "node_id", "ip_address", "tcp_port", "udp_port", "client_name", "client_version",
				"protocol_version", "fork_id", "head_hash", "network_id", "chain_id",
				"first_seen", "last_seen", "country_code", "city_name", "as_number",
				"connection_status", "failure_count", "ping_rtt", "updated_at", "created_at",
			}))

		nodes, err := repo.GetActiveNodes(0)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		// empty slice is acceptable for this case
		if nodes == nil {
			nodes = []*db.Node{} // consistency: nil -> empty slice
		}
		if len(nodes) != 0 {
			t.Errorf("Expected empty nodes slice, got %d nodes", len(nodes))
		}
	})

	t.Run("UpdateNodeStatus with empty nodeID", func(t *testing.T) {
		mock.ExpectExec("UPDATE el_nodes SET connection_status = \\$1, failure_count = \\$2, updated_at = NOW\\(\\) WHERE node_id = \\$3").
			WithArgs(db.NodeStatusOnline, 0, "").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateNodeStatus("", db.NodeStatusOnline, 0)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("GetNodesByClient with empty client name", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM el_nodes WHERE client_name ILIKE \\$1 ORDER BY last_seen DESC LIMIT \\$2 OFFSET \\$3").
			WithArgs("%%", 10, 0).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "node_id", "ip_address", "tcp_port", "udp_port", "client_name", "client_version",
				"protocol_version", "fork_id", "head_hash", "network_id", "chain_id",
				"first_seen", "last_seen", "country_code", "city_name", "as_number",
				"connection_status", "failure_count", "ping_rtt", "updated_at", "created_at",
			}))

		nodes, err := repo.GetNodesByClient("", 10, 0)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		// empty slice is acceptable for this case
		if nodes == nil {
			nodes = []*db.Node{} // consistency: nil -> empty slice
		}
		if len(nodes) != 0 {
			t.Errorf("Expected empty nodes slice, got %d nodes", len(nodes))
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestNodeRepository_DistributionMethods(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	database := &db.Database{DB: mockDB}
	repo := db.NewNodeRepository(database)

	t.Run("GetClientDistribution", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"client_name", "count"}).
			AddRow("Geth", 45).
			AddRow("Nethermind", 30).
			AddRow("Besu", 15).
			AddRow("Erigon", 10)

		mock.ExpectQuery("SELECT client_name, COUNT\\(\\*\\) as count FROM el_nodes WHERE client_name IS NOT NULL GROUP BY client_name ORDER BY count DESC").
			WillReturnRows(mockRows)

		distribution, err := repo.GetClientDistribution()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := map[string]int64{
			"Geth":       45,
			"Nethermind": 30,
			"Besu":       15,
			"Erigon":     10,
		}

		for client, expectedCount := range expected {
			if actualCount, exists := distribution[client]; !exists {
				t.Errorf("Expected client %s not found", client)
			} else if actualCount != expectedCount {
				t.Errorf("Expected count %d for %s, got %d", expectedCount, client, actualCount)
			}
		}
	})

	t.Run("GetCountryDistribution", func(t *testing.T) {
		mockRows := sqlmock.NewRows([]string{"country_code", "count"}).
			AddRow("US", 50).
			AddRow("DE", 25).
			AddRow("CN", 15).
			AddRow("JP", 10)

		mock.ExpectQuery("SELECT country_code, COUNT\\(\\*\\) as count FROM el_nodes WHERE country_code IS NOT NULL GROUP BY country_code ORDER BY count DESC").
			WillReturnRows(mockRows)

		distribution, err := repo.GetCountryDistribution()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := map[string]int64{
			"US": 50,
			"DE": 25,
			"CN": 15,
			"JP": 10,
		}

		for country, expectedCount := range expected {
			if actualCount, exists := distribution[country]; !exists {
				t.Errorf("Expected country %s not found", country)
			} else if actualCount != expectedCount {
				t.Errorf("Expected count %d for %s, got %d", expectedCount, country, actualCount)
			}
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestNodeUpsert_FieldValidation(t *testing.T) {
	tests := []struct {
		name  string
		node  db.NodeUpsert
		valid bool
	}{
		{
			name: "valid node with all fields",
			node: db.NodeUpsert{
				NodeID:           "valid-node-id",
				IPAddress:        "192.168.1.1",
				TCPPort:          30303,
				UDPPort:          30303,
				ClientName:       "Geth",
				ClientVersion:    "1.10.26",
				ProtocolVersion:  67,
				ForkID:           stringPtr("0x123"),
				HeadHash:         stringPtr("0xabc"),
				NetworkID:        int64Ptr(1),
				ChainID:          int64Ptr(1),
				LastSeen:         time.Now(),
				FirstSeen:        time.Now(),
				CountryCode:      stringPtr("US"),
				CityName:         stringPtr("New York"),
				ASNumber:         int64Ptr(15169),
				ConnectionStatus: db.NodeStatusOnline,
				FailureCount:     0,
				PingRTT:          int64Ptr(50000),
			},
			valid: true,
		},
		{
			name: "minimal valid node",
			node: db.NodeUpsert{
				NodeID:           "minimal-node-id",
				IPAddress:        "10.0.0.1",
				TCPPort:          30303,
				UDPPort:          30303,
				ClientName:       "Unknown",
				ClientVersion:    "0.0.0",
				ProtocolVersion:  0,
				LastSeen:         time.Now(),
				FirstSeen:        time.Now(),
				ConnectionStatus: db.NodeStatusUnknown,
				FailureCount:     0,
			},
			valid: true,
		},
		{
			name: "empty node ID",
			node: db.NodeUpsert{
				NodeID:           "",
				IPAddress:        "192.168.1.1",
				TCPPort:          30303,
				UDPPort:          30303,
				ClientName:       "Geth",
				ClientVersion:    "1.10.26",
				ProtocolVersion:  67,
				ConnectionStatus: db.NodeStatusOnline,
			},
			valid: false, // nodeid should not be empty in practice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// field accessibility and basic validation
			if tt.valid {
				if tt.node.NodeID == "" && tt.name != "empty node ID" {
					t.Error("Valid node should have non-empty NodeID")
				}
				if tt.node.IPAddress == "" {
					t.Error("Valid node should have non-empty IPAddress")
				}
				if tt.node.TCPPort < 0 || tt.node.TCPPort > 65535 {
					t.Error("TCP port should be in valid range")
				}
				if tt.node.UDPPort < 0 || tt.node.UDPPort > 65535 {
					t.Error("UDP port should be in valid range")
				}
			}
		})
	}
}

func TestNode_FieldTypes(t *testing.T) {
	node := db.Node{
		ID:               1,
		NodeID:           "test-node-id",
		IPAddress:        "192.168.1.1",
		TCPPort:          30303,
		UDPPort:          30303,
		ClientName:       "Geth",
		ClientVersion:    "1.10.26",
		ProtocolVersion:  67,
		ForkID:           sql.NullString{String: "0x123", Valid: true},
		HeadHash:         sql.NullString{String: "0xabc", Valid: true},
		NetworkID:        sql.NullInt64{Int64: 1, Valid: true},
		ChainID:          sql.NullInt64{Int64: 1, Valid: true},
		FirstSeen:        time.Now(),
		LastSeen:         time.Now(),
		CountryCode:      sql.NullString{String: "US", Valid: true},
		CityName:         sql.NullString{String: "New York", Valid: true},
		ASNumber:         sql.NullInt64{Int64: 15169, Valid: true},
		ConnectionStatus: db.NodeStatusOnline,
		FailureCount:     0,
		PingRTT:          sql.NullInt64{Int64: 50000, Valid: true},
		UpdatedAt:        time.Now(),
		CreatedAt:        time.Now(),
	}

	// test that all fields are accessible and have expected types
	if node.ID <= 0 {
		t.Error("Node ID should be positive")
	}
	if !node.ForkID.Valid || node.ForkID.String != "0x123" {
		t.Error("ForkID not set correctly")
	}
	if !node.HeadHash.Valid || node.HeadHash.String != "0xabc" {
		t.Error("HeadHash not set correctly")
	}
	if !node.NetworkID.Valid || node.NetworkID.Int64 != 1 {
		t.Error("NetworkID not set correctly")
	}
	if !node.ChainID.Valid || node.ChainID.Int64 != 1 {
		t.Error("ChainID not set correctly")
	}
	if !node.CountryCode.Valid || node.CountryCode.String != "US" {
		t.Error("CountryCode not set correctly")
	}
	if !node.CityName.Valid || node.CityName.String != "New York" {
		t.Error("CityName not set correctly")
	}
	if !node.ASNumber.Valid || node.ASNumber.Int64 != 15169 {
		t.Error("ASNumber not set correctly")
	}
	if !node.PingRTT.Valid || node.PingRTT.Int64 != 50000 {
		t.Error("PingRTT not set correctly")
	}
	if node.ConnectionStatus != db.NodeStatusOnline {
		t.Error("ConnectionStatus not set correctly")
	}
}

// Helper functions for creating pointers to basic types (1/2).
func stringPtr(s string) *string {
	return &s
}

// Helper functions for creating pointers to basic types (2/2).
func int64Ptr(i int64) *int64 {
	return &i
}
