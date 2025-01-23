package main

import (
	"errors"
	"fmt"
	"net/http"

	common "github.com/rodrigueghenda/commons"
	pb "github.com/rodrigueghenda/commons/api"
	"github.com/rodrigueghenda/omsv2-gateway/gateway"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc/codes"
	otelCodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	gateway gateway.OrdersGateway
}

func NewHandler (gateway gateway.OrdersGateway) *handler {
	return &handler{gateway}
}
func (h *handler) registerRoutes(mux *http.ServeMux) {
	// static folder serving
	mux.Handle("/", http.FileServer(http.Dir("public")))

	mux.HandleFunc("POST /api/customers/{customerID}/orders", h.handleCreateOrder)
	mux.HandleFunc("GET /api/customers/{customerID}/orders/{orderID}", h.handleGetOrder)
}




//function to pass through customer ID & order ID to get  specific order
func (h *handler) handleGetOrder ( w http.ResponseWriter, r http.Request) {
	customerID := r.PathValue("customerID")
	orderID := r.PathValue("orderID")

	tr := otel.Tracer("http")
	ctx, span := tr.Start(r.Context(), fmt.Sprint("%s %s", r.Method, 
	r.RequestURI))
	defer span.End()

	//fetching order using orderID & customerID
	o, err := h.gateway.GetOrder(ctx, orderID,customerID)
	rStatus := status.Convert(err)
	
	//GRPC error handling message
	if rStatus != nil {
		span.SetStatus(otelCodes.Error, err.Error())
	  if rStatus.Code() != codes.InvalidArgument {
		common.WriteError(w, http.StatusBadRequest, rStatus.Message())
			return
		}

		common.WriteJSON(w, http.StatusOK, o)
	}

}
 


func (h *handler) handleCreateOrder ( w http.ResponseWriter, r http.Request) {
  customerID := r.PathValue("customerID")

  //Slice of items we can send to our Grpc server 
  var items []*pb.ItemsWithQuantity

  //check "r" statement might cause issues
  if err := common.READJSON(&r , &items); err != nil {
	// Error message
	common.WriteError(w, http.StatusBadRequest, err.Error())
	return
  }

  tr := otel.Tracer("http")
  ctx, span := tr.Start(r.Context(), fmt.Sprint("%s %s", r.Method, 
  r.RequestURI))
  defer span.End()

  //A function that validates the input
  if err := validateItems(items); err != nil {
  // Error message
	common.WriteError(w, http.StatusBadRequest, err.Error())
	return
  }

  o, err := h.gateway.CreateOrder(ctx, &pb.CreateOrderRequest{
	CustomerID: customerID,
	Items: items,
  })

  //GRPC error handling message
  rStatus := status.Convert(err)
  if rStatus != nil {
	span.SetStatus(otelCodes.Error, err.Error())
	if rStatus.Code() != codes.InvalidArgument {
		common.WriteError(w, http.StatusBadRequest, rStatus.Message())
		return
	}
  }

  // Error handling message
  if err != nil {
	common.WriteError(w, http.StatusInternalServerError, err.Error())
	return
  }

  res := &CreateOrderRequest{
	Order: 		o,
	RedirectToURL: fmt.Sprintf("http://localhost:8080/success.html?customerID=%s&orderID=%s", o.CustomerID, o.ID),
  }
  
  //return message to user
  common.WriteJSON(w, http.StatusOK, o)
}



 //A function that validates the user input
func validateItems(Items []*pb.ItemsWithQuantity) error {

	//if length of item is "0" return an error
 if len(Items) == 0 {
	return common.ErrNoItems
 }

 //loop over the items to check if its an empty string
 for _, i := range Items {
	if i.ID == "" {
		return errors.New("item ID is required")
	}

// if items quantity is equal to "0" 	
	if i.Quantity <= 0 {
		return errors.New("items must have valid quantity")
	} 
 }
 return nil
}