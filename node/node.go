package main

import (
	pb "auctionSystem/grpc"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const auctionDone = 200

type Node struct {
	pb.UnimplementedAuctionServer
	mu      sync.Mutex
	lamport int64
	winner  bidInstance
	peers   []pb.AuctionClient
	id      int64
}

type bidInstance struct {
	amount   int64
	lamport  int64
	clientId string
}

func (n *Node) Bid(ctx context.Context, req *pb.BidRequest) (*pb.BidReply, error) {
	n.mu.Lock()
	//update time
	n.lamport = max(n.lamport, req.Lamport)
	n.tick()

	now := n.lamport

	//check if over
	if n.auctionOver() {
		n.mu.Unlock()
		return &pb.BidReply{Status: pb.Status_ACTIONOVER, Lamport: now}, nil
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
	currentBid.lamport++
	n.winner = currentBid
	n.mu.Unlock()

	n.Broadcast(currentBid)

	return &pb.BidReply{Status: pb.Status_SUCCESS, Lamport: now}, nil
}

func (n *Node) Result(ctx context.Context, req *pb.ResultRequest) (*pb.ResultReply, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.lamport = max(n.lamport, req.Lamport)
	n.tick()

	return &pb.ResultReply{
		AuctionIsOver: n.auctionOver(),
		HighestBid:    n.winner.amount,
		HighestBidder: n.winner.clientId,
		Lamport:       n.lamport,
	}, nil
}

func (n *Node) connectOthers(ports []int64) error {
	for _, p := range ports {
		if p == n.id {
			continue
		}
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
	updated := &pb.StateUpdate{
		ClientId: currentBid.clientId,
		Amount:   currentBid.amount,
		Lamport:  currentBid.lamport,
	}

	var wg sync.WaitGroup
	for _, peer := range n.peers {

		wg.Add(1)
		go func(p pb.AuctionClient) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if _, err := p.Replicate(ctx, updated); err != nil {
				log.Printf("Node couldnt replicate to peer: %v", err)
			}
		}(peer)
	}
	wg.Wait()
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

func main() {
	port := flag.Int64("port", 5000, "self port")
	peers := flag.String("peers", "", "other nodes, seperate by comma")
	flag.Parse()

	ports, err := parsePorts(*peers)
	if err != nil {
		log.Fatalf("bad -peers: %v", err)
	}

	node := &Node{
		id: *port,
	}

	node.tick()

	if err := node.connectOthers(ports); err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterAuctionServer(server, node)

	if err := server.Serve(lis); err != nil {
		log.Fatalf("server: %v", err)
	}
}

func parsePorts(s string) ([]int64, error) {
	var ports []int64
	for _, field := range strings.Split(s, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		port, err := strconv.ParseInt(field, 10, 64)
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}
	return ports, nil
}

func (n *Node) Replicate(ctx context.Context, req *pb.StateUpdate) (*pb.Ack, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.lamport = max(n.lamport, req.Lamport)
	n.tick()

	incoming := bidInstance{amount: req.Amount, lamport: req.Lamport, clientId: req.ClientId}
	if incoming.beats(n.winner) {
		n.winner = incoming
	}
	return &pb.Ack{Lamport: n.lamport}, nil
}

func (n *Node) tick() {
	n.lamport++
}
