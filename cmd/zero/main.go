package main

import (
	pb "example.com/graphd/cmd/zero/grpc"
	"fmt"
	"github.com/boltdb/bolt"
	"go.uber.org/zap"
	"log"
	"net"
)

const (
	GRPCAddr = "localhost:4448"
	HTTPAddr = "localhost:4447"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	logger.Info("Hello from zap logger")
	// this map contains the mapping from groupId to nodeId
	// we can use this to check out if a node is down in each group.
	// each group would coordinate using raft
	// as long as the majority of nodes in each group are live we would have consistency

	handle, err := bolt.Open("./build/data/master/", 0600, bolt.DefaultOptions)
	if err != nil {
		logger.Fatal("Could not open connection to bolt storage")
		return
	}

	ch, err := newConsistentHashHandler(handle)

	listener, err := net.Listen("tcp", GRPCAddr)
	if err != nil {
		log.Fatalf("Error in starting Zero GRPC Listener: %v\n", err)
	}
	zeroServer, err := newZeroServer(handle, logger)
	pb.RegisterZeroServer(zeroServer.Server, zeroServer)
	err = zeroServer.Server.Serve(listener)
	if err != nil {
		logger.Fatal("Could not start zero GRPC server")
		return
	}

	httpSrv := &httpService{
		addr:   HTTPAddr,
		logger: logger,
		server: zeroServer,
		c:      ch,
	}
	logger.Info(fmt.Sprintf("Running Zero at addr: %s, %s", HTTPAddr, GRPCAddr))
	httpSrv.Start()
}
