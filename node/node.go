package main

import (
	pb "auctionSystem/grpc"
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const auctionDone = 100

type Node struct {
	pb.UnimplementedAuctionServer
	mu      sync.Mutex
	lamport int64
	winner  bidInstance
	peers   []pb.AuctionClient
}

type bidInstance struct {
	amount   int64
	lamport  int64
	clientId string
}

func (n *Node) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidReply, error) {
	n.mu.Lock()
	//update time
	n.lamport = max(n.lamport, req.lamport) + 1
	now := n.lamport

	//check if over
	if n.auctionOver() {
		n.mu.Unlock()
		return &pb.BidReply{Status: pb, pb.status_ACTIONVER, lamport: now}, nil
	}

	//check if valid bid
	if req.ClientId == "" || req.Amount <= 0 {
		n.mu.Unlock()
		return &pb.BidReply{Status: pb.Status_EXCEPTION, Lamport: now}, nil
	}

	// make instance of bid
	currentBid := bidInstance{amount: req.Amount, lamport: now, clientId: req.ClientId}

	//check if beats
	if !currentBid.beats(n.winner) {
		n.mu.Unlock()
		return &pb.BidReply{Status: pb.Status_FAIL, Lamport: now}, nil
	}

	//assume here it beats
	n.winner = currentBid
	n.mu.Unlock()

	n.Broadcast(currentBid)

	return &pb.BidReply{Status: pb.Status_SUCCESS, Lamport: now}, nil
}

func (n *Node) Result(ctx context.Context, req *pb.ResultRequest) (*pb.ResultReply, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	resReply := &pb.ResultReply{}
	// Auction is over
	if n.auctionOver() {
		resReply.AuctionIsOver = true
	} else {
		resReply.AuctionIsOver = false
	}
	n.lamport++
	resReply.HighestBidder = n.winner.clientId
	resReply.HighestBid = n.winner.amount
	return nil, ctx.Err()
}

func (n *Node) connectOthers(ports []int64) error {
	for _, p := range ports {
		conn, err := grpc.NewClient(
			fmt.Sprintf("localhost:%d", p),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return err
		}
		n.peers = append(n.peers, pb.NewAuctionClient(conn))
	}
	return nil
}

func (n *Node) Broadcast(currentBid bidInstance) {
	for _, other := range n.peers {

	}
}

func (b bidInstance) beats(cur bidInstance) bool {
	if cur.clientId == "" {
		return true
	}
	if b.amount != cur.amount {
		return b.amount > cur.amount
	}
	if b.lamport != cur.lamport {
		return b.lamport < cur.lamport
	}
	return b.clientId < cur.clientId
}

func (n *Node) auctionOver() bool {
	return n.lamport > auctionDone
}
