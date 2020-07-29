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
type autentification struct{
  Access string
  Refresh string
}


func main() {
  client :=  Connect()

}




func Connect() mongo.client {
  // Create client
  fmt.Println("Create client")
  client, err := mongo.NewClient(options.Client().ApplyURI("mongodb+srv://user:user@clustervs.rwrgh.mongodb.net/BaseOne?retryWrites=true&w=majority"))
  if err != nil {
    log.Fatal(err)
  }


  // Create connect
  fmt.Println("Create connect")
  err = client.Connect(context.TODO())
  if err != nil {
    log.Fatal(err)
  }


  // Check the connection
  fmt.Println("Check the connection")
  err = client.Ping(context.TODO(), nil)
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println("Connected to MongoDB Succesfull")

  return client


}

func addData(){
  user1 := autentification{"GhkjhsKd89DhjivsusHhfuidh9fvhu1","lilkdjfnlHLIUHerfgiwebllsdbLJDBLK9fvhu1"}
  user2 := autentification{"GhkjhsKd89DhjivsusHhfuidh9fvhu2","lilkdjfnlHLIUHerfgiwebllsdbLJDBLK9fvhu2"}
  user3 := autentification{"GhkjhsKd89DhjivsusHhfuidh9fvhu3","lilkdjfnlHLIUHerfgiwebllsdbLJDBLK9fvhu3"}

  collection := client.Database("BaseOne").Collection("ACol")
  fmt.Println("Connected to Datebase and Collection!")

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

  fmt.Println(collection.FindOne(context.TODO(), nil))
}


func close(){
  err := client.Disconnect(context.TODO())
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println("Connection to MongoDB closed.")
}
