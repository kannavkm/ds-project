package main

import (
	"flag"
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
	"strings"
	"time"
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

	id := flag.Int("id", 0, "Id of the cluster")
	port := flag.String("p", "8001", "Port to listen on")
	clusterS := flag.String("cluster", "", "Comma Separated ips of the rafts of current nodes")
	flag.Parse()
	if len(*port) != 4 {
		fmt.Println("Usage server [-p] port ...")
		flag.PrintDefaults()
		os.Exit(1)
	}
	var cluster []string = strings.Split(*clusterS, ",")
	for i := range cluster {
		cluster[i] = strings.Trim(cluster[i], " ")
	}
	logger.Info(fmt.Sprintf("Running Node: %d at port: %s", *id, *port))
}
