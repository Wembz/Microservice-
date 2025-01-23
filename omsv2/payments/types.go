package payments

import ("context"
pb "github.com/rodrigueghenda/commons/api"
)

// defining a set of methods related to payment services in the future.
type PaymentsService interface {
	CreatePayment(context.Context, *pb.Order) (string, error)
}