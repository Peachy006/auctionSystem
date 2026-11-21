package main

import (
	"context"
	"fmt"
	"log"

	as "auctionSystem/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	name    string
	nodes   []as.AuctionClient // list of servers
	lamport int64
}

func NewClient(name string, ids []int) (*Client, error) {
	client := &Client{name: name}

	for _, id := range ids {
		addr := fmt.Sprintf("localhost:%d", 1000+id)
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, err
		}
		client.nodes = append(client.nodes, as.NewAuctionClient(conn))
	}
	return client, nil
}

func (c *Client) Bid(amount int) {
	c.lamport++
	req := &as.BidRequest{ClientId: c.name, Amount: int64(amount), Lamport: c.lamport}

	for i, node := range c.nodes {
		reply, err := node.Bid(context.Background(), req)
		if err != nil {
			log.Printf("Error sending bid to node %d: %v", i+1, err)
			continue
		}

		c.updateLamport(reply.Lamport)
		log.Printf("Bid sent to node %d: Status=%s, Lamport=%d", i+1, reply.Status, reply.Lamport)
		return
	}
	log.Printf("Failed to send bid to any node")
}

func (c *Client) Result() {
	c.lamport++
	req := &as.ResultRequest{Lamport: c.lamport}

	for i, node := range c.nodes {
		reply, err := node.Result(context.Background(), req)
		if err != nil {
			log.Printf("Error requesting result from node %d: %v", i+1, err)
			continue
		}

		c.updateLamport(reply.Lamport)
		if reply.AuctionIsOver {
			log.Printf("Auction is over. Winner: %s with bid: %d", reply.winner, reply.HighestBid)
		} else {
			log.Printf("highest bid: %d by %s", reply.HighestBid, reply.winner)
		}
		return
	}
	log.Printf("Failed to get result from any node")
}

func (c *Client) updateLamport(received int64) {
	if received > c.lamport {
		c.lamport = received
	}
	c.lamport++
}
