package main

import (
	"context"
	"fmt"
	

	pb "github.com/rodrigueghenda/commons/api"
	"go.opentelemetry.io/otel/trace"
)

type TelemetryMiddleware struct {
	next OrderService
}

func NewTelemetryMiddleware(next OrderService) OrderService {
	return &TelemetryMiddleware{next}
}

//Function to GET order 
func (s *TelemetryMiddleware) GetOrder (ctx context.Context, p *pb.GetOrderRequest)(*pb.Order, error){
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("GetOrder: %v", p))

	return s.next.GetOrder(ctx, p)
}

func (s *TelemetryMiddleware) UpdateOrder(ctx context.Context, o *pb.Order) (*pb.Order, 
	error) {
		span := trace.SpanFromContext(ctx)
		span.AddEvent(fmt.Sprintf("UpdateOrder: %v", o))


	return s.next.UpdateOrder(ctx, o)	
	}
//fucntion to created a customer order
func (s *TelemetryMiddleware) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest, items []*pb.Item) (*pb.Order, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("GetOrder: %v", p))


	return s.next.CreateOrder(ctx, p, items)
}
 
//function to create a slice of order created by user 
func (s *TelemetryMiddleware) ValidateOrder(ctx context.Context, p  *pb.CreateOrderRequest)([]*pb.Item, error){
	span := trace.SpanFromContext(ctx)
	span.AddEvent(fmt.Sprintf("ValidateOrder: %v", p))
	
	return s.next.ValidateOrder(ctx, p)
}


