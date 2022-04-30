package main

import (
	"github.com/buraksezer/consistent"
	"go.uber.org/zap"
	"log"
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
	cfg := consistent.Config{
		// groups are the members of the ring
		// each key can map to
		PartitionCount:    271,
		ReplicationFactor: 20,
		Load:              1.25,
		Hasher:            hasher{},
	}

	c := consistent.New(nil, cfg)
	ch := consistentHashHandler{}

}
