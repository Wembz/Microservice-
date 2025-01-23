package gateway

import (
	"context"

	pb "github.com/rodrigueghenda/commons/api"
)

type KitchenGateway interface {
	UpdateOrder(context.Context, *pb.Order) error
}