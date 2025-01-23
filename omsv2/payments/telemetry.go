package payments

import (
	"context"
	"fmt"

	pb "github.com/rodrigueghenda/commons/api"
	"go.opentelemetry.io/otel/trace"
)

type TelemetryMiddleware struct{
	next PaymentsService
}

func NewTelemetryMiddleware (next PaymentsService) PaymentsService {
	return &TelemetryMiddleware{next}
}

// How to create customer Payment function
func (s *TelemetryMiddleware) CreatePayment(ctx context.Context, o *pb.Order) (string, error) {
	span := trace.SpanContextFromContext(ctx)
	span.AddEvent(fmt.Sprintf("CreatePayment: %v", o))
	
	return s.next.CreatePayment(ctx, o)
}