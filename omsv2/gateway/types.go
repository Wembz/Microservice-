package main

import pb "github.com/rodrigueghenda/commons/api"

type CreateOrderRequest struct {
	Order         *pb.Order `"json": order`
	RedirectToURL string    `"json": redirectToURL`
}