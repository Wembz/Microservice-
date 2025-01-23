package payments

import (
	"context"
	"testing"

	"github.com/rodrigueghenda/commons/api"
	inmemRegistry "github.com/rodrigueghenda/commons/discovery/inmem"
	"github.com/rodrigueghenda/omsv2-payments/gateway"
	"github.com/rodrigueghenda/omsv2-payments/processor/inmem"
)

func TestService(t *testing.T) {
	processor := inmem.NewInmem()
	registry := inmemRegistry.NewRegistry()

	gateway := gateway.NewGateway(registry)
	svc := NewService(processor, gateway)

	t.Run("should create a payment link", func(t *testing.T) {
		link, err := svc.CreatePayment(context.Background(), &api.Order{})
		if err != nil {
			t.Errorf("CreatePayment() error = %v, want nil", err)
		}

		if link == "" {
			t.Error("CreatePayment() link is empty")
		}
	})
}