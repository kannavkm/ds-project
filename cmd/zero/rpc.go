package main

import (
	"context"
	pb "example.com/graphd/cmd/zero/grpc"
	"github.com/hashicorp/go-uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math"
	"sync"
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
	mut    sync.Mutex
	gInfo  map[string]*groupInfo
	nInfo  map[string]*pb.Node
	Server *grpc.Server
	logger *zap.Logger
	c      *consistentHashHandler
	pb.UnimplementedZeroServer
}

func newZeroServer(logger *zap.Logger, ch *consistentHashHandler) (*ZeroServer, error) {
	return &ZeroServer{
		gInfo:  make(map[string]*groupInfo),
		nInfo:  make(map[string]*pb.Node),
		Server: grpc.NewServer(),
		logger: logger,
		c:      ch,
	}, nil
}

func (z *ZeroServer) CreateAGroup(ctx context.Context, node *pb.Node) (*pb.Group, error) {
	z.logger.Info("node asking to create a group")
	group := groupInfo{
		leader:  node,
		members: 1,
	}
	uid, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}

	z.mut.Lock()
	defer z.mut.Unlock()
	err = z.c.addGroup(uid)
	if err != nil {
		return nil, err
	}

	node.GroupId = uid
	z.nInfo[node.GetId()] = node
	z.gInfo[uid] = &group

	return &pb.Group{
		Id:                uid,
		LeaderRaftAddress: node.GetRaftAddress(),
		LeaderHttpAddress: node.GetHttpAddress(),
		Members:           1,
	}, nil
}

func (z *ZeroServer) JoinAGroup(ctx context.Context, node *pb.Node) (*pb.Group, error) {
	z.logger.Info("node asking to join a group")
	// if we already know about this group then
	z.mut.Lock()
	defer z.mut.Unlock()
	if memNode, ok := z.nInfo[node.GetId()]; ok {
		ginfo := z.gInfo[memNode.GetGroupId()]
		leader := ginfo.leader
		return &pb.Group{
			Id:                memNode.GetGroupId(),
			LeaderRaftAddress: leader.GetRaftAddress(),
			LeaderHttpAddress: leader.GetHttpAddress(),
			Members:           int32(ginfo.members), // this would not change
			// else we do not know about this guy
		}, nil
	}
	minMemberGroup := ""
	minMembers := math.MaxUint32
	for k, v := range z.gInfo {
		if int(v.members) < minMembers {
			minMemberGroup = k
			minMembers = int(v.members)
		}
	}

	z.logger.Info("Group selected finally", zap.String("name", minMemberGroup))
	if minMemberGroup == "" {
		z.logger.Error("Could not find minimum member group")
		return nil, nil
	}
	if entry, ok := z.gInfo[minMemberGroup]; ok {
		entry.members += 1
		node.GroupId = minMemberGroup
		z.nInfo[node.GetId()] = node
	}

	return &pb.Group{
		Id:                minMemberGroup,
		LeaderRaftAddress: z.gInfo[minMemberGroup].leader.GetRaftAddress(),
		LeaderHttpAddress: z.gInfo[minMemberGroup].leader.GetRaftAddress(),
		Members:           int32(z.gInfo[minMemberGroup].members),
	}, nil
}

func (z *ZeroServer) UpdateLeader(ctx context.Context, node *pb.Node) (*pb.Group, error) {
	z.logger.Info("node asking to update leader of group")
	// update leader
	z.mut.Lock()
	defer z.mut.Unlock()
	var memN int
	if entry, ok := z.gInfo[node.GetGroupId()]; ok {
		z.nInfo[node.GetId()] = node
		entry.leader = node
		memN = entry.members
	}
	return &pb.Group{
		Id:                node.GetGroupId(),
		LeaderRaftAddress: node.GetRaftAddress(),
		LeaderHttpAddress: node.GetHttpAddress(),
		Members:           int32(memN),
	}, nil
}

func (z *ZeroServer) GetGroupInfo(id string) (*pb.Group, error) {

	z.mut.Lock()
	defer z.mut.Unlock()
	var memN int
	var leaderHTTP, leaderRaft string
	if entry, ok := z.gInfo[id]; ok {
		leaderHTTP = entry.leader.GetHttpAddress()
		leaderRaft = entry.leader.GetRaftAddress()
		memN = entry.members
	}
	return &pb.Group{
		Id:                id,
		LeaderRaftAddress: leaderRaft,
		LeaderHttpAddress: leaderHTTP,
		Members:           int32(memN),
	}, nil
}

func (z *ZeroServer) GetLeader(context.Context, *pb.Group) (*pb.Node, error) {
	z.logger.Info("node asking to fetch the leader of a group")
	return nil, nil
}
