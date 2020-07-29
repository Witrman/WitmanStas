package main

import (
	"context"
	"fmt"
	"log"
	//  "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"go.mongodb.org/mongo-driver/mongo/readpref"
)

type autentification struct {
	GUID    string
	Access  string
	Refresh string
}

var client *mongo.Client

func main() {
	ConnectDB()
	addDataDB()
	closeDB()
}

func ConnectDB() {
	// Create client
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb+srv://user:user@clustervs.rwrgh.mongodb.net/BaseOne?retryWrites=true&w=majority"))
	if err != nil {
		log.Fatal(err)
	}

	// Create connect
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB Succesfull")
}

func addDataDB() {
	user1 := autentification{"Jdfjvnsskdjvns1", "GhkjhsKd89DhjivsusHhfuidh9fvhu1", "lilkdjfnlHLIUHerfgiwebllsdbLJDBLK9fvhu1"}
	user2 := autentification{"Jdfjvnsskdjvns2", "GhkjhsKd89DhjivsusHhfuidh9fvhu2", "lilkdjfnlHLIUHerfgiwebllsdbLJDBLK9fvhu2"}
	user3 := autentification{"Jdfjvnsskdjvns3", "GhkjhsKd89DhjivsusHhfuidh9fvhu3", "lilkdjfnlHLIUHerfgiwebllsdbLJDBLK9fvhu3"}

	collection := client.Database("BaseOne").Collection("ACol")
	fmt.Println("Connected to Datebase and Collection!")
	fmt.Println(collection.FindOne(context.TODO(), user1))
	err := collection.Drop(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	insertResult, err := collection.InsertOne(context.TODO(), user1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	users := []interface{}{user2, user3}
	insertManyResult, err := collection.InsertMany(context.TODO(), users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)

}

func closeDB() {
	err := client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
}
