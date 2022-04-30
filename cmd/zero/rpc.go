package main

import (
	"context"
	pb "example.com/graphd/cmd/zero/grpc"
	"github.com/boltdb/bolt"
	"github.com/hashicorp/go-uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math"
	"strconv"
)

var (
	leaderHTTPAddress = []byte("LeaderHTTPAddress")
	leaderRaftAddress = []byte("LeaderRaftAddress")
	memberNum         = []byte("MembersNumber")
)

// zero accepts rpc responses from the multiple alpha nodes
// and Also sends rpc requests to appropriate alpha nodes

type groupInfo struct {
	leader  *pb.Node
	members int
}

// ZeroServer this handles the grpc request to the zero
// this might need to be exported
type ZeroServer struct {
	db     *bolt.DB
	Server *grpc.Server
	logger *zap.Logger
	pb.UnimplementedZeroServer
}

func newZeroServer(db *bolt.DB, logger *zap.Logger) (*ZeroServer, error) {
	tx, err := db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Create all the buckets
	if _, err := tx.CreateBucketIfNotExists(leaderHTTPAddress); err != nil {
		return nil, err
	}
	if _, err := tx.CreateBucketIfNotExists(leaderRaftAddress); err != nil {
		return nil, err
	}
	// Create all the buckets
	if _, err := tx.CreateBucketIfNotExists(memberNum); err != nil {
		return nil, err
	}
	tx.Commit()
	return &ZeroServer{
		db:     db,
		Server: grpc.NewServer(),
		logger: logger,
	}, nil
}

func (g *ZeroServer) CreateAGroup(ctx context.Context, node *pb.Node) (*pb.Group, error) {
	g.logger.Info("node asking to create a group")
	txn, err := g.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()
	uid, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}
	bucket := txn.Bucket(leaderHTTPAddress)
	if err := bucket.Put([]byte(uid), []byte(node.GetHttpAddress())); err != nil {
		return nil, err
	}
	bucket = txn.Bucket(leaderRaftAddress)
	if err := bucket.Put([]byte(uid), []byte(node.GetRaftAddress())); err != nil {
		return nil, err
	}
	bucket = txn.Bucket(memberNum)
	if err := bucket.Put([]byte(uid), []byte(strconv.FormatInt(1, 10))); err != nil {
		return nil, err
	}
	txn.Commit()
	return &pb.Group{
		Id:                uid,
		LeaderRaftAddress: node.GetRaftAddress(),
		LeaderHttpAddress: node.GetHttpAddress(),
		Members:           1,
	}, nil
}

func (g *ZeroServer) JoinAGroup(ctx context.Context, node *pb.Node) (*pb.Group, error) {
	g.logger.Info("node asking to join a group")
	minMemberGroup, err := func() (string, error) {
		txn, err := g.db.Begin(false)
		if err != nil {
			return "", err
		}
		defer txn.Rollback()
		minMemberGroup := ""
		minMembers := math.MaxUint32
		err = txn.Bucket(memberNum).ForEach(func(k, v []byte) error {
			groupid := string(k)
			num, err := strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				return err
			}
			if int(num) < minMembers {
				minMemberGroup = groupid
				minMembers = int(num)
			}
			return nil
		})
		txn.Commit()
		return minMemberGroup, err
	}()
	g.logger.Info("Group selected finally", zap.String("name", minMemberGroup))
	if err != nil {
		g.logger.Error("Could not find minimum member group", zap.Error(err))
		return nil, err
	}
	txn, err := g.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()
	bucket := txn.Bucket(leaderHTTPAddress)
	if err != nil {
		return nil, err
	}
	leaderHTTP := string(bucket.Get([]byte(minMemberGroup)))
	bucket = txn.Bucket(leaderRaftAddress)
	if err != nil {
		return nil, err
	}
	leaderRaft := string(bucket.Get([]byte(minMemberGroup)))
	bucket = txn.Bucket(memberNum)
	if err != nil {
		return nil, err
	}
	members := bucket.Get([]byte(minMemberGroup))
	memN, err := strconv.ParseInt(string(members), 10, 64)
	if err != nil {
		return nil, err
	}
	if err := bucket.Put([]byte(minMemberGroup), []byte(strconv.FormatInt(memN+1, 10))); err != nil {
		return nil, err
	}
	txn.Commit()
	return &pb.Group{
		Id:                minMemberGroup,
		LeaderRaftAddress: leaderRaft,
		LeaderHttpAddress: leaderHTTP,
		Members:           int32(memN + 1),
	}, nil
}

func (g *ZeroServer) UpdateLeader(ctx context.Context, node *pb.Node) (*pb.Group, error) {
	g.logger.Info("node asking to update leader of group")
	txn, err := g.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()
	bucket := txn.Bucket(leaderHTTPAddress)
	if err != nil {
		return nil, err
	}
	if err := bucket.Put([]byte(node.GetGroupId()), []byte(node.GetHttpAddress())); err != nil {
		return nil, err
	}
	bucket = txn.Bucket(leaderRaftAddress)
	if err != nil {
		return nil, err
	}
	if err := bucket.Put([]byte(node.GetGroupId()), []byte(node.GetRaftAddress())); err != nil {
		return nil, err
	}
	members := bucket.Get([]byte(node.GetGroupId()))
	memN, err := strconv.ParseInt(string(members), 10, 64)
	txn.Commit()
	return &pb.Group{
		Id:                node.GetGroupId(),
		LeaderRaftAddress: node.GetRaftAddress(),
		LeaderHttpAddress: node.GetHttpAddress(),
		Members:           int32(memN),
	}, nil
}

func (g *ZeroServer) GetGroupInfo(id string) (*pb.Group, error) {
	txn, err := g.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()
	bucket := txn.Bucket(leaderHTTPAddress)
	if err != nil {
		return nil, err
	}
	leaderHTTP := string(bucket.Get([]byte(id)))
	bucket = txn.Bucket(leaderRaftAddress)
	if err != nil {
		return nil, err
	}
	leaderRaft := string(bucket.Get([]byte(id)))
	bucket = txn.Bucket(memberNum)
	if err != nil {
		return nil, err
	}
	members := bucket.Get([]byte(id))
	memN, err := strconv.ParseInt(string(members), 10, 64)
	return &pb.Group{
		Id:                id,
		LeaderRaftAddress: leaderRaft,
		LeaderHttpAddress: leaderHTTP,
		Members:           int32(memN),
	}, nil
}

func (g *ZeroServer) GetLeader(context.Context, *pb.Group) (*pb.Node, error) {
	g.logger.Info("node asking to fetch the leader of a group")
	return nil, nil
}
