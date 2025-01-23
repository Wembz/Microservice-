package gateway

import ("context"
pb "github.com/rodrigueghenda/commons/api"
)

//function that will connect us to the other services

type OrdersGateway interface {
  CreateOrder (context.Context, *pb.CreateOrderRequest) (*pb.Order, error)
  GetOrder (ctx context.Context, orderID, CustomerID string) (*pb.Order, error)

}