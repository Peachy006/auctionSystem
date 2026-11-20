package main

import (
	pb "auctionSystem/grpc"
	"context"
	"sync"
)

type Node struct {
	pb.UnimplementedAuctionServer
	mu            sync.Mutex
	lamport       int64
	highestBid    int64
	highestBidder string
	otherPorts    []int64
	otherNodes    []pb.AuctionServer
}

func main() {
	//Hej
}

func (n *Node) BidRequest(ctx context.Context, req *pb.BidRequest) (*pb.BidReply, error) {
	n.mu.Lock()
	reply := &pb.BidReply{}
	defer n.mu.Lock()
	if req.Amount <= n.highestBid {
		reply.Status = 2
		return reply, ctx.Err()
	}

	return nil, ctx.Err()
}

func (n *Node) Broadcast(req *pb.BidRequest) {
	for _, other := range otherNode {

	}
}
