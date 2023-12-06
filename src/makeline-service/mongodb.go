package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBOrderRepo struct {
	db *mongo.Collection
}

func NewMongoDBOrderRepo() *MongoDBOrderRepo {
	// Get database uri from environment variable
	mongoUri := os.Getenv("ORDER_DB_URI")
	if mongoUri == "" {
		log.Printf("ORDER_DB_URI is not set")
		return nil
	}

	// get database name from environment variable
	mongoDb := os.Getenv("ORDER_DB_NAME")
	if mongoDb == "" {
		log.Printf("ORDER_DB_NAME is not set")
		return nil
	}

	// get database collection name from environment variable
	mongoCollection := os.Getenv("ORDER_DB_COLLECTION_NAME")
	if mongoCollection == "" {
		log.Printf("ORDER_DB_COLLECTION_NAME is not set")
		return nil
	}

	// get database username from environment variable
	mongoUser := os.Getenv("ORDER_DB_USERNAME")

	// get database password from environment variable
	mongoPassword := os.Getenv("ORDER_DB_PASSWORD")

	// create a context
	ctx := context.Background()

	// create a mongo client
	var clientOptions *options.ClientOptions
	if mongoUser == "" && mongoPassword == "" {
		clientOptions = options.Client().ApplyURI(mongoUri)
	} else {
		clientOptions = options.Client().ApplyURI(mongoUri).
			SetAuth(options.Credential{
				AuthSource: mongoDb,
				Username:   mongoUser,
				Password:   mongoPassword,
			}).
			SetTLSConfig(&tls.Config{InsecureSkipVerify: false})
	}

	mongoClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Printf("failed to connect to mongodb: %s", err)
		return nil
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Printf("failed to ping database: %s", err)
	} else {
		log.Printf("pong from database")
	}

	// get a handle for the collection
	collection := mongoClient.Database(mongoDb).Collection(mongoCollection)
	//defer collection.Database().Client().Disconnect(context.Background())

	return &MongoDBOrderRepo{collection}
}

func (r *MongoDBOrderRepo) GetPendingOrders() ([]Order, error) {
	ctx := context.TODO()

	var orders []Order
	cursor, err := r.db.Find(ctx, bson.M{"status": Pending})
	if err != nil {
		log.Printf("Failed to find records: %s", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	// Check if there was an error during iteration
	if err := cursor.Err(); err != nil {
		log.Printf("Failed to find records: %s", err)
		return nil, err
	}

	// Iterate over the cursor and decode each document
	for cursor.Next(ctx) {
		var pendingOrder Order
		if err := cursor.Decode(&pendingOrder); err != nil {
			log.Printf("Failed to decode order: %s", err)
			return nil, err
		}
		orders = append(orders, pendingOrder)
	}

	return orders, nil
}

func (r *MongoDBOrderRepo) GetOrder(id string) (Order, error) {
	var ctx = context.TODO()

	singleResult := r.db.FindOne(ctx, bson.M{"orderid": id})
	var order Order
	err := singleResult.Decode(&order)
	if err != nil {
		log.Printf("Failed to decode order: %s", err)
		return order, err
	}

	return order, nil
}

func (r *MongoDBOrderRepo) InsertOrders(orders []Order) error {
	ctx := context.TODO()

	var ordersInterface []interface{}
	for _, o := range orders {
		ordersInterface = append(ordersInterface, interface{}(o))
	}

	if len(ordersInterface) == 0 {
		log.Printf("No orders to insert into database")
	} else {
		// Insert orders
		insertResult, err := r.db.InsertMany(ctx, ordersInterface)
		if err != nil {
			log.Printf("Failed to insert order: %s", err)
			return err
		}

		log.Printf("Inserted %v documents into database\n", len(insertResult.InsertedIDs))
	}
	return nil
}

func (r *MongoDBOrderRepo) UpdateOrder(order Order) error {
	var ctx = context.TODO()

	log.Printf("Updating order: %v", order)

	// Update the order
	updateResult, err := r.db.UpdateMany(
		ctx,
		bson.M{"orderid": order.OrderID},
		bson.D{
			{Key: "$set", Value: bson.D{{Key: "status", Value: order.Status}}},
		},
	)
	if err != nil {
		log.Printf("Failed to update order: %s", err)
		return err
	}

	log.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	return nil
}
