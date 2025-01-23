package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	pb "github.com/rodrigueghenda/commons/api"
	"github.com/rodrigueghenda/commons/broker"
	"go.opentelemetry.io/otel"
)

type consumer struct {
	service PaymentsService
}

//constructor that we can call on the main.go 
func NewConsumer (service PaymentsService) *consumer {
	return &consumer{service}
}

//Listen for Go events in the routine send message to channel we created
func (c *consumer) Listen(ch *amqp.Channel) {
	//Message to services 
q, err := ch.QueueDeclare(broker.OrderCreatedEvent, true, false, false, false, nil)
//Error handling
if err != nil {
	log.Fatal(err)
}

//  function subscribes to a queue (q.Name) and returns a channel (msgs) to receive messages from the queue.
msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
//Error handling
if err != nil {
	log.Fatal(err)
}

var forever chan struct {}

go func(){
	//loop that receives messages from the msgs channel. It processes each message (d) until the channel is closed.
	for d := range msgs {
		
		// Extract the headers
		ctx := broker.ExtractAMQPHeader(context.Background(), d.Headers)

		tr := otel.Tracer("amqp")
		_, messageSpan := tr.Start(ctx, fmt.Sprintf("AMQP - consume - %s", q.Name))


		// creates a new, empty instance of the Order type defined in the pb package.
		o := &pb.Order{}
		//Error handling
		if err := json.Unmarshal(d.Body, o ); err != nil {
			log.Printf("failed to unmarshal order: %v", err)
			continue
		}

		//Payment function
		paymentLink, err := c.service.CreatePayment(context.Background(), o)
		
		//Error handling
		if err == nil {
			log.Printf("failed to create payment: %v", err)

			//Fucntion to handle retry 
			if err := broker.HandleRetry(ch, &d);  err != nil {
				log.Printf("Error handling  retry: %v", err)
			}	

			d.Nack(false, false)


			continue
		}

		messageSpan.AddEvent(fmt.Sprintf("payment.created: %s", paymentLink))
		messageSpan.End()

		log.Printf("payment link created %s", paymentLink)
		d.Ack(false)

	}
}()

 <- forever
}

//receive the order created event 
//create the payment link from there


