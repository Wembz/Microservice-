package main

import (
	"context"
	"log"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"
	common "github.com/rodrigueghenda/commons"
	"github.com/rodrigueghenda/commons/discovery"
	"github.com/rodrigueghenda/commons/discovery/consul"
	"github.com/rodrigueghenda/omsv2-gateway/gateway"
)

var (
	serviceName = "gateway"
	httpAddr = common.EnvString("HTTP_ADDR", "8080")
	consulAddr = common.EnvString("CONSUL_ADDR", "localhost:2000")
	jaegerAddr = common.EnvString("JAEGAR_ADDR", "localhost: 4318")
	//connectiong to the order service
	//orderServiceAddr = "localhost:2000"
)


//Gateway acts as pathway to HTTP SERVER whilst implementing Grpc into the service
func main() {
	if err := common.SetGlobalTracer(context.TODO(), serviceName, jaegerAddr)
	err != nil {
		log.Fatal("failed to set global tracer")
	}


  registry, err := consul.NewRegistry(consulAddr, serviceName)
  if err != nil {
	panic(err)

  }

  ctx := context.Background()
  instanceID := discovery.GenerateInstanceID(serviceName)
  if err := registry.Register(ctx, instanceID, serviceName, httpAddr); err != nil {
	panic(err)
  }

   //Healthcheck status 
   go func ()  {
	for {
		if err := registry.HealthCheck(instanceID, serviceName); err != nil {
			log.Fatal("failed to health check")
		}
		time.Sleep(time.Second * 1)
	}
  }()

  defer registry.Deregister(ctx, instanceID, serviceName)

	mux := http.NewServeMux()

	OrdersGateway := gateway.NewGRPCGateway(registry)
	handler := NewHandler(OrdersGateway)
	handler.registerRoutes(mux)

	//notification server is Working
	log.Printf("Starting HTTP serverat %s", httpAddr)
	
	//notification server has failed
	if err := http.ListenAndServe(httpAddr, mux); err != nil {
		log.Fatal("Failed to start http server")
	}

}