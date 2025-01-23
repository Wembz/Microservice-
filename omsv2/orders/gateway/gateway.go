package gateway

import ( "context"
pb "github.com/rodrigueghenda/commons/api"
)

type StockGateway interface {
	CheckIfItemIsInStock(ctx context.Context, customerID string, items []*pb.ItemsWithQuantity) (bool, []*pb.Item, error)
}


// var itemsWithPrice []*pb.Item
// for _, i := range mergedItems {
	//itemsWithPrice = append(itemsWithPrice, &pb.Item{
	//	PriceID: "price_1QcrOJRtfBUIej9CeYn9VfqN",
	//	ID: i.ID,
	//	Quantity: i.Quantity,
//	})
 //}
