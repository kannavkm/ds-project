package main

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"go.uber.org/zap"
	"net"
)

type config struct {
	port string `yaml:"port"`
}

// The full server encapsulated in a struct
type serverNode struct {
	logger  *zap.Logger // logger
	address net.Addr    // address to run the http server
	raft    *raft.Raft  // the raft
	fsm     *raftFSM    // the fsm
}

func newServer(config *config, logger *zap.Logger) *serverNode {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		logger.Fatal("Could not open connection to badger db", zap.Error(err))
	}
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = config.id
	//raftConfig.Logger
	transport, err := raft.NewTCPTransport()
	fsm := raftFSM{db: db, logger: logger}
	server := serverNode{logger: logger, address: config.address}
}
