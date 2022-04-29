package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"go.uber.org/zap"
)

type config struct {
	id   string
	path string
	addr string
}

// The full server encapsulated in a struct
type server struct {
	cfg    *config
	logger *zap.Logger // logger
	raft   *raft.Raft  // the raft
	fsm    *raftFSM    // the fsm
	db     *badger.DB
}

func (s *server) get(key uint64) (uint64, error) {
	keyS := strconv.FormatUint(key, 10)
	var valS uint64

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(keyS))

		err = item.Value(func(val []byte) error {
			valS, err = strconv.ParseUint(string(val), 10, 64)
			return err
		})

		return err
	})
	if err != nil {
		log.Fatal(err)
		return 0, err
	}
	return valS, err
}

func (s *server) put(key, val uint64) error {

	data := event{
		opType: "SET",
		key:    key,
		value:  val,
	}

	dataJson, err := json.Marshal(data)

	if err != nil {
		log.Fatal(err)
	}

	applyFuture := s.raft.Apply(dataJson, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		log.Fatal(err)
	}

	type ResData struct {
		Error error
		Data  interface{}
	}

	_, ok := applyFuture.Response().(*ResData)

	if !ok {
		log.Fatal("Invalid Response")
	}
	return nil
}

func (s *server) delete(key uint64) error {

	data := event{
		opType: "SET",
		key:    key,
		value:  0,
	}

	dataJson, err := json.Marshal(data)

	if err != nil {
		log.Fatal(err)
	}

	applyFuture := s.raft.Apply(dataJson, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		log.Fatal(err)
	}

	type ResData struct {
		Error error
		Data  interface{}
	}

	_, ok := applyFuture.Response().(*ResData)

	if !ok {
		// TODO
	}
	return nil
}

func (s *server) join(joinAddr, id string) error {
	b, err := json.Marshal(map[string]string{"addr": s.cfg.addr, "id": id})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func newServer(cfg *config, logger *zap.Logger) (*server, error) {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		logger.Fatal("Could not open connection to badger db", zap.Error(err))
	}

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(cfg.id)

	if err != nil {
		return nil, err
	}
	snapshots, err := raft.NewFileSnapshotStore(cfg.path, retainSnapshotCount, os.Stderr)
	if err != nil {
		return nil, err
	}
	transport, err := raft.NewTCPTransport(cfg.addr, cfg.addr, 3, raftTimeout, os.Stderr)
	if err != nil {
		return nil, err
	}
	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(cfg.path, "raft.db"))
	if err != nil {
		return nil, err
	}
	logStore := boltDB
	stableStore := boltDB
	fsm := raftFSM{db: db, logger: logger}
	rf, err := raft.NewRaft(raftConfig, &fsm, logStore, stableStore, snapshots, transport)
	srv := &server{
		logger: logger,
		raft:   rf,
		fsm:    &fsm,
		db:     db,
		cfg:    cfg,
	}
	return srv, nil
}
