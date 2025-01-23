package main

import (
	"context"

	pb "github.com/rodrigueghenda/commons/api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DbName   = "orders"
	CollName = "orders"
)

// Slice ready to store orders as they are added
var orders = make([]*pb.Order, 0)

type store struct {
	db *mongo.Client
}

func NewStore(db *mongo.Client) *store {
	return &store{db}

}

// function to CREATE New order
func (s *store) Create(ctx context.Context, o Order) (primitive.ObjectID, error) {
	col := s.db.Database(DbName).Collection(CollName)

	newOrder, err := col.InsertOne(ctx, o)

	id := newOrder.InsertedID.(primitive.ObjectID)
	return id, err
}

// function to GET order
func (s *store) Get(ctx context.Context, id, customerID string) (*Order, error) {
	col := s.db.Database(DbName).Collection(CollName)

	oID, _ := primitive.ObjectIDFromHex(id)

	var o Order
	err := col.FindOne(ctx, bson.M{
		"_id":        oID,
		"customerID": customerID,
	}).Decode(&o)

	// Error message
	return &o, err
}

func (s *store) Update(ctx context.Context, id string, newOrder *pb.Order) error {
	col := s.db.Database(DbName).Collection(CollName)

	oID, _ := primitive.ObjectIDFromHex(id)

	_, err := col.UpdateOne(ctx,
		bson.M{"_id": oID},
		bson.M{"$set": bson.M{
			"paymentlink": newOrder.PaymentLink,
			"status":      newOrder.Status, 
		}})

	return err	

		 
}
