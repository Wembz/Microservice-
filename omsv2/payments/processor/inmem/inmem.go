package inmem

import (
	pb "github.com/rodrigueghenda/commons/api"
)

type Inmem struct {}

func NewInmem() *Inmem {
	return &Inmem{}
}
//dummy link to create payment link 
func (i *Inmem) CreatePaymentLink(o *pb.Order) (string, error) {
	return "dummy-link", nil
}