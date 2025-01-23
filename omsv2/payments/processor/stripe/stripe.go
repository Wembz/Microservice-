package stripe

import (
	"fmt"
	"log"

	common "github.com/rodrigueghenda/commons"
	pb "github.com/rodrigueghenda/commons/api"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
)

// variable for the http server
var gatewayHTTPAddr = common.EnvString("GATEWAY_HTTP_ADDRESS", "http://localhost:8080")

type Stripe struct{}

// CreatePaymentLink implements processor.PaymentProcessor.
func (s *Stripe) CreatePaymentLink(*pb.Order) (string, error) {
	panic("unimplemented")
}

func NewProcessor() *Stripe {
	return &Stripe{}
}

// function to create payment checkout link for user/customer
func CreatePaymentLink(o *pb.Order) (string, error) {
	log.Printf("Creating payment link for order %v", o)

	gatewaySuccessURL := fmt.Sprintf("%s/success.html?customerID=%s&orderID=%s", gatewayHTTPAddr, o.CustomerID, o.ID)
	gatewayCancelURL := fmt.Sprintf("%s/cancel.html", gatewayHTTPAddr, )

	// var for items using slice & for loop
	items := []*stripe.CheckoutSessionLineItemParams{}
	for _, item := range o.Items {
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			Price:    stripe.String(item.PriceID),
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}

	params := &stripe.CheckoutSessionParams{
		Metadata: map[string]string{
			"orderID": o.ID,
			"customerID": o.CustomerID,
		},
		LineItems:  items,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(gatewaySuccessURL),
		CancelURL:  stripe.String(gatewayCancelURL),
	}

	result, err := session.New(params)
	// Error handling
	if err != nil {
		return "", nil
	}

	return result.URL, nil

}
