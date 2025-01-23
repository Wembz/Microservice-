package main

import (
	"context"
	
	common "github.com/rodrigueghenda/commons"
	pb "github.com/rodrigueghenda/commons/api"
	"github.com/rodrigueghenda/omsv2-orders/gateway"
)

type Service struct {
	store OrdersStore
	gateway gateway.StockGateway
}

func NewService(store OrdersStore, gateway gateway.StockGateway) *Service {
	return &Service{store, gateway}
}
//Function to GET order 
func (s *Service) GetOrder (ctx context.Context, p *pb.GetOrderRequest)(*pb.Order, error){
	o, err := s.store.Get(ctx, p.OrderID, p.CustomerID)
	if err != nil {
		return nil, err 
	}

	return o.ToProto(), nil
}

func (s *Service) UpdateOrder(ctx context.Context, o *pb.Order) (*pb.Order, 
	error) {
		err :=  s.store.Update(ctx, o.ID, o)
		if err != nil {
			return nil, err
		}
		return o, nil
	}
//fucntion to created a customer order
func (s *Service) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest, items []*pb.Item) (*pb.Order, error) {
	
	//Function to call the createOrder service 
	id, err := s.store.Create(ctx, Order{
		CustomerID:  p.CustomerID,
		Status:      "pending",
		Items:       items,
		PaymentLink: "",
	})
	if err != nil {
		return nil, err
	}

	o := &pb.Order{
		ID:         id.Hex(),
		CustomerID: p.CustomerID,
		Status:     "pending",
		Items:      items,
	}
  return o, nil
}
 
//function to create a slice of order created by user 
func (s *Service) ValidateOrder(ctx context.Context, p  *pb.CreateOrderRequest)([]*pb.Item, error){
 if len(p.Items) == 0 {
	return nil, common.ErrNoItems
 }

 mergedItems := mergeItemsQuantities(p.Items)
 

 // validate with the stock service
inStock, items, err := s.gateway.CheckIfItemIsInStock(ctx, p.
CustomerID, mergedItems)
if err != nil {
	return nil, err
}

if !inStock {
	return items, common.ErrNoStock
}

 return items, nil
} 

// How to merge items 
func mergeItemsQuantities(Items []*pb.ItemsWithQuantity) []*pb.ItemsWithQuantity {
	merged := make([]*pb.ItemsWithQuantity, 0 )

	//for loop to merge items
	for _, item := range Items {
		found := false
		for _, finalItem := range merged {
			if finalItem.ID == item.ID {
			  finalItem.Quantity += item.Quantity
			  found = true
			  break	
			}
		}

		if !found {
			merged = append(merged, item)
		}

	}
	return merged
}