package broker

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

const MaxRetrycount = 3 
const DQL = "dlq_main"

//function connects to server
func Connect(user, pass, host, port string) (*amqp.Channel, func()error ) {
	address := fmt.Sprintf("amqp//%s:%s@%s:%s", user, pass, host, port)

	conn, err := amqp.Dial(address)
	if err != nil {
		log.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	// function communicates with server and sends message that order has been created
	err = ch.ExchangeDeclare(OrderCreatedEvent, "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}
	// function communicates with server and sends message to multiple services
	err = ch.ExchangeDeclare(OrderPaidEvent, "fanout", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = createDLQAndDLX(ch)
	if err != nil {
		log.Fatal(err)
	}

	return ch, conn.Close
}

func HandleRetry(ch *amqp.Channel, d *amqp.Delivery) error {
	if d.Headers ==  nil {
		d.Headers = amqp.Table{}
	}

	// implentation to retry count 
	retryCount, ok := d.Headers["X-retry-count"]. (int64)
	if !ok {
		retryCount = 0 
	}

	retryCount++
	d.Headers["X-retry-count"] = retryCount

	log.Printf("Retrying message %s, retry count: %d", d.Body, retryCount)

	if retryCount >= MaxRetrycount{
		//DLQ
		log.Printf("Moving messageto DLQ %s", DLQ)

		return ch.PublishWithContext(context.Background(), "", DLQ, false, false, amqp.Publishing{
			ContentType: "application/json",
			Headers: d.Headers,
			Body: d.Body,
			DeliveryMode: amqp.Persistent,
		})
	}

	time.Sleep(time.Second * time.Duration(retryCount))

	return ch.PublishWithContext(
	context.Background(),
	d.Exchange,
	d.RoutingKey,
	false,
	false,
	amqp.Publishing{
		ContentType: "application/json",
		Headers: d.Headers,
		Body: d.Body,
		DeliveryMode: amqp.Persistent,
	},
	)	
}

//function to intialize all the functions
func createDLQAndDLX(ch *amqp.Channel) error {
	q, err := ch.QueueDeclare(
		"main_queue", //name
		true, //durable
		false, // delete when unused
		false, //exclusive 
		false, // no-wait
		nil, //arugments
	)
	if err != nil {
		return err
	}
	// Declare DLX 
	dlx := "dlx_main"
	err = ch.ExchangeDeclare(
	  dlx,      // name
	  "fanout", // type
	  true,     // durable
	  false,    // auto-declared
	  false,    // internal 
	  false,    // no-wait
	  nil,      // arguements
	)

	if err !=nil {
		return err
	}

	//Bind  main queue to DLX 
	err = ch.QueueBind(
	  q.Name, //queue name
	  "",     //routing key
	  dlx, 	  //exchange
	  false,     
	  nil, 			
	)
	if err != nil {
		return err
	}

	//Declare DLQ
	_, err = ch.QueueDeclare(
	  DLQ,     // name 
	  true,    // durable 
	  false,   // delete when unused 
	  false,   // exclusive	
	  false,   // no wait
	  nil,     // arguments 
	)
	if err != nil {
		return err 
	}

	return err

}

type AmqpHeaderCarrier map[string]interface{}

func (a AmqpHeaderCarrier ) Get(k string) string {
  value, ok := a[k]
  if !ok {
	return ""
  }	

  return value.(string)
}

func (a AmqpHeaderCarrier ) Set(k string, v string) {
  a[k] = v
}

func (a AmqpHeaderCarrier )Keys() [] string {
  keys := make([]string, len(a))
  i := 0

  for k := range a {
	keys[i] = k 
	i++
  }

  return keys
}

//Injection
func InjectAMQPHeaders(ctx context.Context) map[string] interface{}{
	carrier := make(AmqpHeaderCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier )
	return carrier
}

//Extraction 
func ExtractAMQPHeader(ctx context.Context, headers map[string] interface{}) context.Context{
	return otel.GetTextMapPropagator().Extract(ctx, AmqpHeaderCarrier(headers))
}