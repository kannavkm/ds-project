package main

import (
	pb "example.com/graphd/cmd/zero/grpc"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"go.uber.org/zap"
	"log"
	"net"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	logger.Info("Hello from zap logger")

	httpAddr := flag.String("haddr", "localhost:4447", "Set the address for the HTTP server")
	grpcAddr := flag.String("gaddr", "localhost:4448", "Set the address for the Raft")

	flag.Parse()

	// we can use this to check out if a node is down in each group.
	// each group would coordinate using raft
	// as long as the majority of nodes in each group are live we would have consistency
	handle, err := bolt.Open("./build/data/master/shard.db", 0600, bolt.DefaultOptions)
	if err != nil {
		fmt.Println(err)
		logger.Fatal("Could not open connection to bolt storage")
		return
	}
	defer handle.Close()
	ch, err := newConsistentHashHandler(handle)

	listener, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		log.Fatalf("Error in starting Zero GRPC Listener: %v\n", err)
	}
	zeroServer, err := newZeroServer(logger, ch)
	pb.RegisterZeroServer(zeroServer.Server, zeroServer)
	httpSrv := &httpService{
		addr:   *httpAddr,
		logger: logger,
		server: zeroServer,
		c:      ch,
	}
	logger.Info(fmt.Sprintf("Running Zero at addr: %s, %s", *httpAddr, *grpcAddr))
	go httpSrv.Start()
	err = zeroServer.Server.Serve(listener)
	if err != nil {
		logger.Fatal("Could not start zero GRPC server")
		return
	}
}
