package processor

import(
	pb "github.com/rodrigueghenda/commons/api"
)

type PaymentProcessor interface {
	CreatePaymentLink(*pb.Order) (string, error)
}

