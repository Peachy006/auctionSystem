package main

import (
	pb "auctionSystem/grpc"
	"context"
	"log"
	"sync"

	"google.golang.org/protobuf/types/known/emptypb"
)

type Node struct {
	pb.UnimplementedAuctionServer
	mu sync.Mutex
}

func main() {
	log.Printf("Sean Seaner")

}

func (n *Node) BidRequest(ctx context.Context, req *pb.BidRequest) (*emptypb.Empty, error) {

}
