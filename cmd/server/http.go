package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"strconv"

	badger "github.com/dgraph-io/badger/v3"
	"go.uber.org/zap"
)

// methods
// 1. join (dynamically adding nodes should be a thing)
// 2. GET /key
// 3. PUT /key
// 4. DELETE /key

func (server *serverNode) handleKeyGet(w http.ResponseWriter, r *http.Request) {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *badger.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	keys := r.URL.Query()["key"]
	key := keys[0]

	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))

		item.Value(func(val []byte) error {
			w.Write(val)
			return nil
		})

		return err
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (server *serverNode) handleKeyPut(w http.ResponseWriter, r *http.Request) {
	keys := r.URL.Query()["key"]
	vals := r.URL.Query()["val"]

	key, _ := strconv.Atoi(keys[0])
	val, _ := strconv.Atoi(vals[0])

	type ReqData struct {
		opType string
		key    uint64
		value  uint64
	}

	data := ReqData{
		opType: "SET",
		key:    uint64(key),
		value:  uint64(val),
	}

	dataJson, err := json.Marshal(data)

	if err != nil {
		log.Fatal(err)
	}

	applyFuture := server.raft.Apply(dataJson, 500*time.Millisecond)
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
}

func (server *serverNode) handleKeyDelete(w http.ResponseWriter, r *http.Request) {
}

func (server *serverNode) Start() {
	log.Fatal(http.ListenAndServe(server.address.String(), nil))
	server.logger.Info("Server Starting", zap.String("address", server.address.String()))

}
