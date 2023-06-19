package database

import (
	"auth-service/constants"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//var CNX = Connection()

func Connection() *mongo.Client {
	// Set client options
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	dbUrl := "mongodb+srv://" + constants.DB_USERNAME + ":" + constants.DB_PASSWORD + "@cluster0.gdbqwc7.mongodb.net/?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(dbUrl).SetServerAPIOptions(serverAPI)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	return client
}
func CloseClientDB(client *mongo.Client) {
	if client == nil {
		return
	}

	err := client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// TODO optional you can log your closed MongoDB client
	fmt.Println("Connection to MongoDB closed.")
}
