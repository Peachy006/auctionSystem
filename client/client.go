package main

import (
	"log"

	as "auctionSystem/grpc"
)

type Client struct {
	name          string
	nodes         []as.AuctionClient //List of node ports
	nodeTalkingTo int
	lamport       int
}

func NewClient(name string, ids []int) (*Client, error) {

	for _, id := range ids {
		log.Printf("localhost:%d", 1000+id)
	}
}

func (c *Client) Bid(amount int) {
	c.lamport++
}

func (c *Client) Result() {
	c.lamport++
}
