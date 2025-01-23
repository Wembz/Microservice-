package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/rodrigueghenda/commons/api"
	"github.com/rodrigueghenda/commons/broker"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"
	"go.opentelemetry.io/otel"
)

type PaymentHTTPHandler struct {
	channel *amqp.Channel
}

func NewPaymentHTTPHandler(channel *amqp.Channel) *PaymentHTTPHandler {
	return &PaymentHTTPHandler{channel}
}

func (h *PaymentHTTPHandler) registerRoutes(router *http.ServeMux) {
	router.HandleFunc("/webhook", h.handleCheckoutWebhook)
}

// function handles incoming webhook requests for a checkout.
func (h *PaymentHTTPHandler) handleCheckoutWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	body, err := io.ReadAll(r.Body)
	//Error handling
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading request body: %v\n", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// prints the content of the request body to the console.
	fmt.Fprintf(os.Stdout, "Got body: %s\n", body)

	//verifies a webhook request from Stripe

	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"),
		endpointStripeSecret)

	//Error handling
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying webhook signature: %v\n", err)
		w.WriteHeader(http.StatusBadRequest) // Return a 400 error on a bad signature
		return
	}

	// the function handles Stripe's checkout.session.completed event
	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing webhook JSON: %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if session.PaymentStatus == "paid" {
			log.Printf("Payment for Checkout Session %v succeeded!", session.ID)

			orderID := session.Metadata["orderID"]
			customerID := session.Metadata["customerID"]

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			o := &pb.Order{
				ID:          orderID,
				CustomerID:  customerID,
				Status:      "paid",
				PaymentLink: "",
			}

			marshalledOrder, err := json.Marshal(o)
			if err != nil {
				log.Fatal(err.Error())
			}

			tr := otel.Tracer("amqp")
			amqpContext, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - publish - %s",
				broker.OrderPaidEvent))
			defer messageSpan.End()

			headers := broker.InjectAMQPHeaders(amqpContext)

			// fanning out message  to broker saying order has been paid
			h.channel.PublishWithContext(amqpContext, broker.OrderPaidEvent, "", false,
				false, amqp.Publishing{

					ContentType:  "application/json",
					Body:         marshalledOrder,
					DeliveryMode: amqp.Persistent,
					Headers:      headers,
				})

			log.Println("Message published order.paid")
		}
	}

	w.WriteHeader(http.StatusOK)
}
