package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

func join(joinAddr, myAddr, id string) error {
	b, err := json.Marshal(map[string]string{"addr": myAddr, "id": id})
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

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	logger.Info("Hello from zap logger")

	id := flag.String("id", "", "Id of the cluster")
	httpAddr := flag.String("haddr", "localhost:8000", "Set the address for the HTTP server")
	raftAddr := flag.String("raddr", "localhost:9000", "Set the address for the Raft")
	joinAddr := flag.String("join", "", "Set the address for the node to join")

	flag.Parse()

	// haddr, err := net.ResolveTCPAddr("tcp", *httpAddr)
	// if err != nil {
	// 	logger.Fatal("Wrong http addr")
	// 	return
	// }

	// raddr, err := net.ResolveTCPAddr("tcp", *raftAddr)
	// if err != nil {
	// 	logger.Fatal("Wrong raft addr")
	// 	return
	// }

	cfg := config{
		id:     *id,
		path:   "./build/data/" + *id,
		addr:   *raftAddr,
		leader: *joinAddr == "",
	}

	srv, err := newServer(&cfg, logger)
	if err != nil {
		logger.Fatal("Could not start raft server, try deleting the data directory")
		return
	}

	if *joinAddr != "" {
		// If I am not the first one then join them
		_, err := net.ResolveTCPAddr("tcp", *joinAddr)
		if err != nil {
			logger.Fatal("Could not find join address")
			return
		}
		err = join(*joinAddr, *raftAddr, *id)
		if err != nil {
			logger.Fatal("Could not join")
		}
		logger.Info("Able to join?")
	}

	httpsrv := &httpService{
		addr:   *httpAddr,
		store:  srv,
		logger: logger,
	}
	httpsrv.Start()
	logger.Info(fmt.Sprintf("Running Node: %d at addr: %s, %s", *id, *httpAddr, *raftAddr))
}
