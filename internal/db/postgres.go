package db

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/netip"
	"time"

	_ "github.com/lib/pq"

	"github.com/200ug/peerlogger/internal/common"
	"github.com/200ug/peerlogger/internal/util"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p/enr"

	beacon "github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
)

// ETH2 is a SSZ encoded field.
type ETH2 []byte

func (v ETH2) ENRKey() string { return "eth2" }

func UpdateNodes(db *sql.DB, geoipProvider *util.GeoIP, nodes []common.NodeJSON) error {
	log.Info("Writing nodes to database", "nodes", len(nodes))

	now := time.Now()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(
		`INSERT INTO nodes(
			id,
			now,
			client_type,
			pk,
			software_version,
			capabilities,
			network_id,
			fork_id,
			blockheight,
			total_difficulty,
			head_hash,
			ip,
			country,
			city,
			first_seen,
			last_seen,
			seq,
			score,
			conn_type,
			asn
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, n := range nodes {
		info := &common.ClientInfo{}
		if n.Info != nil {
			info = n.Info
		}

		if info.ClientType == "" && n.TooManyPeers {
			info.ClientType = "tmp"
		}
		connType := ""
		var portUDP enr.UDP
		if n.N.Load(&portUDP) == nil {
			connType = "UDP"
		}
		var portTCP enr.TCP
		if n.N.Load(&portTCP) == nil {
			connType = "TCP"
		}
		fid := fmt.Sprintf("Hash: %v, Next %v", info.ForkID.Hash, info.ForkID.Next)

		var eth2 ETH2
		if n.N.Load(&eth2) == nil {
			info.ClientType = "eth2"
			var dat beacon.Eth2Data
			err = dat.Deserialize(codec.NewDecodingReader(bytes.NewReader(eth2), uint64(len(eth2))))
			if err == nil {
				fid = fmt.Sprintf("Hash: %v, Next %v", dat.ForkDigest, dat.NextForkEpoch)
			}
		}
		var caps string
		for _, c := range info.Capabilities {
			caps = fmt.Sprintf("%v, %v", caps, c.String())
		}
		var pk string
		if n.N.Pubkey() != nil {
			pk = fmt.Sprintf("X: %v, Y: %v", n.N.Pubkey().X.String(), n.N.Pubkey().Y.String())
		}

		var country, city string
		var asn uint64

		if geoipProvider != nil {
			addr, parseErr := netip.ParseAddr(n.N.IP().String())
			if parseErr == nil {
				geoData, err := geoipProvider.Lookup(addr)
				if err == nil && geoData != nil {
					if geoData.CountryName != nil {
						country = *geoData.CountryName
					}
					if geoData.CityName != nil {
						city = *geoData.CityName
					}
					if geoData.ASNumber != nil {
						asn = uint64(*geoData.ASNumber)
					}
				}
			}
		}

		_, err = stmt.Exec(
			n.N.ID().String(),
			now,
			info.ClientType,
			pk,
			info.SoftwareVersion,
			caps,
			info.NetworkID,
			fid,
			info.Blockheight,
			info.TotalDifficulty.String(),
			info.HeadHash.String(),
			n.N.IP().String(),
			country,
			city,
			n.FirstResponse,
			n.LastResponse,
			n.Seq,
			n.Score,
			connType,
			asn,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func CreateDB(db *sql.DB) error {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS nodes (
		id              TEXT NOT NULL,
		now             TIMESTAMP NOT NULL,
		client_type     TEXT,
		pk              TEXT,
		software_version TEXT,
		capabilities    TEXT,
		network_id      BIGINT,
		fork_id         TEXT,
		blockheight     TEXT,
		total_difficulty TEXT,
		head_hash       TEXT,
		ip              INET,
		country         TEXT,
		city            TEXT,
		first_seen      TIMESTAMP,
		last_seen       TIMESTAMP,
		seq             BIGINT,
		score           BIGINT,
		conn_type       TEXT,
		asn             BIGINT,
		PRIMARY KEY (id, now)
	);
	DELETE FROM nodes;
	`
	_, err := db.Exec(sqlStmt)
	return err
}
