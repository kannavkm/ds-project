package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"go.uber.org/zap"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	logger.Info("Hello from zap logger")

	id := flag.String("id", "", "Id of the cluster")
	httpAddr := flag.String("haddr", ":8000", "Set the address for the HTTP server")
	raftAddr := flag.String("raddr", ":9000", "Set the address for the Raft")
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
		id:   *id,
		path: "./node/" + *id,
		addr: *raftAddr,
	}

	srv, err := newServer(&cfg, logger)
	if err != nil {
		logger.Fatal("Could not start raft server, try deleting the data directory")
		return
	}

	if *joinAddr != "" {
		join, err := net.ResolveIPAddr("ip", *joinAddr)
		if err != nil {
			logger.Fatal("Could not find join address")
			return
		}
	}

	httpsrv := &httpService{
		addr:   *httpAddr,
		store:  srv,
		logger: logger,
	}
	httpsrv.Start()
	logger.Info(fmt.Sprintf("Running Node: %d at addr: %s, %s", *id, *httpAddr, *raftAddr))
}
