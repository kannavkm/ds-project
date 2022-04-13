package main

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"io"
	"sync"
)

type raftHandler struct {
	raft *raft.Raft
}

func New(raft *raft.Raft) *raftHandler {
	return &raftHandler{raft: raft}
}

type fsm struct {
	mutex sync.Mutex
	db    *badger.DB
}

type event struct {
	Type  string
	Value int
}

func (f *fsm) Apply(log *raft.Log) interface{} {
	var e event
	if err := json.Unmarshal(log.Data, &e); err != nil {
		panic("Failed unmarshalling Log entry, this is a bug")
	}
	switch e.Type {
	case "set":
		f.mutex.Lock()
		defer f.mutex.Unlock()

	}
}

func (*fsm) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

func (*fsm) Restore(io.ReadCloser) error {
	return nil
}

type fsmSnapshot struct {
	data []byte
}

func (*fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}
func (*fsmSnapshot) Release() {

}
