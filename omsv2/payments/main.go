package payments

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
	common "github.com/rodrigueghenda/commons"
	"github.com/rodrigueghenda/commons/broker"
	"github.com/rodrigueghenda/commons/discovery"
	"github.com/rodrigueghenda/commons/discovery/consul"
	"github.com/rodrigueghenda/omsv2-payments/gateway"
	stripeProcessor "github.com/rodrigueghenda/omsv2-payments/processor/stripe"
	"github.com/stripe/stripe-go/v78"
	"google.golang.org/grpc"
)

// import "golang.org/x/sys/windows/svc"

var (
	serviceName          = "payments"
	grpcAddr             = common.EnvString("GRPC_ADDR", "localhost:2001")
	consulAddr           = common.EnvString("CONSUL_ADDR", "localhost:2000")
	amqpUser             = common.EnvString("RABBITMQ_USER", "guest")
	amqpPass             = common.EnvString("RABBITMQ_PASS", "guest")
	amqpHost             = common.EnvString("RABBITMQ_HOST", "localhost")
	amqpPort             = common.EnvString("RABBITMQ_PORT", "5672")
	stripeKey            = common.EnvString("STRIPE_KEY", "")
	httpAddr             = common.EnvString("HTTP_ADDR", "localhost:8081")
	endpointStripeSecret = common.EnvString("STRIPE_ENDPOINT_SECRET", "whsec...")
	jaegerAddr = common.EnvString("JAEGAR_ADDR", "localhost: 4318")
)

// Build Server so we can receive the Orders from customer
func main() {

	if err := common.SetGlobalTracer(context.TODO(), serviceName, jaegerAddr)
	err != nil {
		log.Fatal("failed to set global tracer")
	}

	//Creating a new service order
	registry, err := consul.NewRegistry(consulAddr, serviceName)
	//Error handling
	if err != nil {
		panic(err)

	}

	// passing context through function calls.
	ctx := context.Background()

	// unique identifier(ID) for a service
	instanceID := discovery.GenerateInstanceID(serviceName)
	//Error handling
	if err := registry.Register(ctx, instanceID, serviceName, grpcAddr); err != nil {
		panic(err)
	}

	//Healthcheck status
	go func() {
		for {
			if err := registry.HealthCheck(instanceID, serviceName); err != nil {
				log.Fatal("failed to health check")
			}
			time.Sleep(time.Second * 1)
		}
	}()

	defer registry.Deregister(ctx, instanceID, serviceName)

	// stripe setup connections
	stripe.Key = stripeKey

	//broker Connection
	ch, close := broker.Connect(amqpUser, amqpPass, amqpHost, amqpPort)
	defer func() {
		close()
		ch.Close()
	}()

	//Stripe Server
	stripeProcessor := stripeProcessor.NewProcessor()
	gateway := gateway.NewGateway(registry)
	svc := NewService(stripeProcessor, gateway)
	svcWithTelemetry := NewTelemetryMiddleware(svc) 

	amqpconsumer := NewConsumer(svcWithTelemetry)
	go amqpconsumer.Listen(ch)

	// http server
	mux := http.NewServeMux()

	httpServer := NewPaymentHTTPHandler(ch)
	httpServer.registerRoutes(mux)

	go func() {
		log.Printf("Starting HTTP server at %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, mux); err != nil {
			log.Fatal("failed to start http server")
		}
	}()

	//grpc server
	grpcServer := grpc.NewServer()

	l, err := net.Listen("tcp", grpcAddr)

	//Error message if connection fails
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}
	defer l.Close()

	// message if server connects
	log.Println("GRPC Server Started at", grpcAddr)
	if err := grpcServer.Serve(l); err != nil {
		log.Fatal(err.Error())
	}

}
