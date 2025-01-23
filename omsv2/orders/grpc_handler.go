package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/rodrigueghenda/commons/api"
	"github.com/rodrigueghenda/commons/broker"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
)

type grpcHandler struct {
	pb.UnimplementedOrderServiceServer

	service OrderService
	channel *amqp.Channel
}

func NewGRPCHandler(grpcServer *grpc.Server, service OrderService, channel *amqp.Channel) {
	handler := &grpcHandler{
		service: service,
		channel: channel,
	}
	pb.RegisterOrderServiceServer(grpcServer, handler)
}

func (h *grpcHandler) UpdateOrder(ctx context.Context, p *pb.Order) (*pb.Order, error) {
	return h.service.UpdateOrder(ctx, p)
}

func (h *grpcHandler) GetOrder(ctx context.Context, p *pb.GetOrderRequest) (*pb.Order, error) {
	return h.service.GetOrder(ctx, p)
}

// create Order function
func (h *grpcHandler) CreateOrder(ctx context.Context, p *pb.CreateOrderRequest) (*pb.Order, error) {

	//Message to services
	q, err := h.channel.QueueDeclare(broker.OrderCreatedEvent, true, false, false, false, nil)
	// Error handling
	if err != nil {
		log.Fatal(err)
	}

	//How consumer will receieve context with header 
	tr := otel.Tracer("amqp")
	amqpContext, Messagespan := tr.Start(ctx, fmt.Sprintf("AMQP - publish - %s", q.Name))
	defer Messagespan.End()

	//Validate order
	items, err := h.service.ValidateOrder(amqpContext, p)
	//Error handling
	if err != nil {
		return nil, err
	}

	// function to create customer order
	o, err := h.service.CreateOrder(amqpContext, p, items)
	// Error handling
	if err != nil {
		return nil, err
	}

	// Marshalling customer order
	marshalledOrder, err := json.Marshal(0)
	// Error handling
	if err != nil {
		return nil, err
	}

	// inject the headers
	headers := broker.InjectAMQPHeaders(amqpContext)

	// How to publish message
	h.channel.PublishWithContext(amqpContext, "", q.Name, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         marshalledOrder,
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
	})

	return o, nil
}
