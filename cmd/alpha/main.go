package main

// each node is participant of 1 group
// this is saying that each group can have multiple disjoint workers
// Each group is replicated over 3 alpha nodes
//
// for each group it is part

import (
	"bytes"
	"context"
	"encoding/json"
	pb "example.com/graphd/cmd/zero/grpc"
	"flag"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"net/http"
	"time"
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
	masterAddr := flag.String("master", "localhost:10000", "The address of the master")
	isLeader := flag.Bool("leader", false, "is the current node a raft leader (used for bootstrapping)")

	flag.Parse()

	con, err := grpc.Dial(*masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Could not connect to Master", zap.Error(err))
	}
	defer con.Close()
	c := pb.NewZeroClient(con)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cfg := config{
		id:     *id,
		path:   "./build/data/" + *id,
		addr:   *raftAddr,
		leader: *isLeader,
	}

	srv, err := newServer(&cfg, logger)
	if err != nil {
		logger.Fatal("Could not start raft server, try deleting the data directory")
		return
	}

	node := pb.Node{
		Id:          *id,
		GroupId:     "",
		RaftAddress: *raftAddr,
		HttpAddress: *httpAddr,
	}

	if *isLeader {
		r, err := c.CreateAGroup(ctx, &node)
		if err != nil {
			logger.Fatal("Could not contact master, try restarting")
			return
		}
		node.GroupId = r.GetId()
	} else {
		r, err := c.JoinAGroup(ctx, &node)
		if err != nil {
			logger.Fatal("Could not contact master, try restarting")
			return
		}
		joinAddr := r.GetLeaderHttpAddress()
		node.GroupId = r.GetId()
		// If I am not the first one then join them
		_, err = net.ResolveTCPAddr("tcp", joinAddr)
		if err != nil {
			logger.Fatal("Could not find join address")
			return
		}
		err = join(joinAddr, *raftAddr, *id)
		if err != nil {
			logger.Fatal("Could not join")
		}
	}

	go func() {
		leaderChange := <-srv.raft.LeaderCh()
		log.Println("Sending leader change req")
		if leaderChange {
			_, err := c.UpdateLeader(ctx, &node)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	httpsrv := &httpService{
		addr:   *httpAddr,
		store:  srv,
		logger: logger,
	}
	logger.Info(fmt.Sprintf("Running Node: %d at addr: %s, %s", *id, *httpAddr, *raftAddr))
	httpsrv.Start()
}
