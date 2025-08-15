package db_test

import (
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/200ug/peerlogger/internal/common"
	"github.com/200ug/peerlogger/internal/db"
	"github.com/200ug/peerlogger/internal/util"
)

func TestUpdateNodes(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Mock the transaction and prepared statement
	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO nodes")
	mock.ExpectExec("INSERT INTO nodes").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Create mock GeoIP provider
	geoIP, err := util.NewGeoIP("", "")
	if err != nil {
		t.Fatalf("Failed to create GeoIP provider: %v", err)
	}

	// Create test node data
	privKey, _ := crypto.GenerateKey()
	testIP := net.ParseIP("8.8.8.8")
	testNode := enode.NewV4(&privKey.PublicKey, testIP, 30303, 30303)

	nodes := []common.NodeJSON{
		{
			N:             testNode,
			Seq:           1,
			Score:         10,
			FirstResponse: time.Now(),
			LastResponse:  time.Now(),
			Info: &common.ClientInfo{
				ClientType:      "geth",
				SoftwareVersion: 68,
				NetworkID:       1,
				TotalDifficulty: big.NewInt(1000),
				HeadHash:        ethcommon.HexToHash("0x123"),
			},
		},
	}

	// Test the UpdateNodes function
	err = db.UpdateNodes(mockDB, geoIP, nodes)
	if err != nil {
		t.Errorf("UpdateNodes failed: %v", err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations not met: %v", err)
	}
}

func TestUpdateNodesWithNilGeoIP(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Mock the transaction and prepared statement
	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO nodes")
	mock.ExpectExec("INSERT INTO nodes").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Create test node data
	privKey, _ := crypto.GenerateKey()
	testIP := net.ParseIP("8.8.8.8")
	testNode := enode.NewV4(&privKey.PublicKey, testIP, 30303, 30303)

	nodes := []common.NodeJSON{
		{
			N:             testNode,
			Seq:           1,
			Score:         10,
			FirstResponse: time.Now(),
			LastResponse:  time.Now(),
			Info: &common.ClientInfo{
				ClientType:      "geth",
				SoftwareVersion: 68,
				NetworkID:       1,
				TotalDifficulty: big.NewInt(1000),
				HeadHash:        ethcommon.HexToHash("0x123"),
			},
		},
	}

	// Test with nil GeoIP provider
	err = db.UpdateNodes(mockDB, nil, nodes)
	if err != nil {
		t.Errorf("UpdateNodes failed: %v", err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations not met: %v", err)
	}
}

func TestCreateDB(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	// Mock the CREATE TABLE and DELETE statements
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS nodes").WillReturnResult(sqlmock.NewResult(0, 0))

	err = db.CreateDB(mockDB)
	if err != nil {
		t.Errorf("CreateDB failed: %v", err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations not met: %v", err)
	}
}