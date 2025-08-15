package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// nodes.json file format, holds a set of node records as a JSON object
type NodeSet map[enode.ID]NodeJSON

type NodeJSON struct {
	Seq uint64      `json:"seq"`
	N   *enode.Node `json:"record"`
	// liveness check tracker: incremented by one when node passes a check, halved every time it doesn't
	Score int `json:"score,omitempty"`
	// last successful contacts
	FirstResponse time.Time `json:"firstResponse"`
	LastResponse  time.Time `json:"lastResponse"`
	// last attempt to contact the node
	LastCheck    time.Time   `json:"lastCheck"`
	Info         *ClientInfo `json:"clientInfo,omitempty"`
	TooManyPeers bool        `json:"tooManyPeers,omitempty"`
}

func LoadNodesJSON(file string) NodeSet {
	var nodes NodeSet
	if err := common.LoadJSON(file, &nodes); err != nil {
		panic(err)
	}
	return nodes
}

func (nodes NodeSet) WriteNodesJSON(file string) {
	nodesJSON, err := json.Marshal(nodes)
	if err != nil {
		panic(err)
	}
	if file == "-" {
		os.Stdout.Write(nodesJSON)
		return
	}
	if err := os.WriteFile(file, nodesJSON, 0644); err != nil {
		panic(err)
	}
}

// Nodes returns the node records contained in the set.
func (ns NodeSet) Nodes() []*enode.Node {
	result := make([]*enode.Node, 0, len(ns))
	for _, n := range ns {
		result = append(result, n.N)
	}
	// Sort by ID.
	sort.Slice(result, func(i, j int) bool {
		return bytes.Compare(result[i].ID().Bytes(), result[j].ID().Bytes()) < 0
	})
	return result
}

// Add ensures the given nodes are present in the set.
func (ns NodeSet) Add(nodes ...*enode.Node) {
	for _, n := range nodes {
		v := ns[n.ID()]
		v.N = n
		v.Seq = n.Seq()
		ns[n.ID()] = v
	}
}

// TopN returns the top n nodes by score as a new set.
func (ns NodeSet) TopN(n int) NodeSet {
	if n >= len(ns) {
		return ns
	}

	byScore := make([]NodeJSON, 0, len(ns))
	for _, v := range ns {
		byScore = append(byScore, v)
	}
	sort.Slice(byScore, func(i, j int) bool {
		return byScore[i].Score >= byScore[j].Score
	})
	result := make(NodeSet, n)
	for _, v := range byScore[:n] {
		result[v.N.ID()] = v
	}
	return result
}

// Verify performs integrity checks on the node set.
func (ns NodeSet) Verify() error {
	for id, n := range ns {
		if n.N.ID() != id {
			return fmt.Errorf("invalid node %v: ID does not match ID %v in record", id, n.N.ID())
		}
		if n.N.Seq() != n.Seq {
			return fmt.Errorf("invalid node %v: 'seq' does not match seq %d from record", id, n.N.Seq())
		}
	}
	return nil
}
