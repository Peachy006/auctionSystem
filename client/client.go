package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	as "auctionSystem/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	name    string
	nodes   []as.AuctionClient // list of servers
	lamport int64
}

var currentNode = 0

func NewClient(name string, ids []int) (*Client, error) {
	client := &Client{name: name}

	for _, id := range ids {
		addr := fmt.Sprintf("localhost:%d", id)
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
		if i < currentNode {
			continue
		}
		reply, err := node.Bid(context.Background(), req)
		if err != nil {
			log.Printf("Error sending bid to node %d", i+1)
			currentNode++
			continue
		}

		c.updateLamport(reply.Lamport)
		log.Printf("Bid sent to node %d: Status=%s, Lamport=%d", i+1, reply.Status, reply.Lamport)
		return
	}
	log.Printf("Failed to send bid to any node")
}

func (c *Client) Result() (*as.ResultReply, error) {
	c.lamport++
	req := &as.ResultRequest{Lamport: c.lamport}

	for _, node := range c.nodes {
		reply, err := node.Result(context.Background(), req)
		if err != nil {
			continue
		}
		c.updateLamport(reply.Lamport)
		return reply, nil
	}
	return nil, fmt.Errorf("no node answered")
}

func (c *Client) PrintResult(reply *as.ResultReply) {
	if reply.AuctionIsOver {
		log.Printf("Auction over, the winner is: %s with %d", reply.HighestBidder, reply.HighestBid)
	} else {
		log.Printf("highest bid is %d by %s", reply.HighestBid, reply.HighestBidder)
	}
}

func (c *Client) updateLamport(received int64) {
	if received > c.lamport {
		c.lamport = received
	}
	c.lamport++
}

func main() {
	var name string
	fmt.Print("Enter your name: ")
	if _, err := fmt.Scan(&name); err != nil {
		return
	}

	ids := []int{5000, 5001, 5002, 5003, 5004, 5005, 5006, 5007}

	client, err := NewClient(name, ids)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("You can now bid and see result")
	fmt.Println("To place a bid: 'bid <amount>' to see result of bidding: 'result'")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		switch rand.Intn(5) + 1 {

		case 1:
			reply, err := client.Result()
			if err != nil {
				log.Print(err)
				continue
			}
			if reply.AuctionIsOver {
				client.PrintResult(reply)
				return
			}
			client.Bid(int(reply.HighestBid) + 2)

		case 2: // just look
			reply, err := client.Result()
			if err != nil {
				log.Print(err)
				continue
			}
			client.PrintResult(reply)
			if reply.AuctionIsOver {
				return
			}
		}
	}

}
