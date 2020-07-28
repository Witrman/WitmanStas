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

func Connect(){
  // Create client
  client, err := mongo.NewClient(options.Client().ApplyURI("witmanstas-shard-00-00.ei6tk.mongodb.net:27017"))
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

  fmt.Println("Connected to MongoDB!")

  collection := client.Database("WitmanBase").Collection("AutCol")
  fmt.Println("Connected to Datebase and Collection!")


}


func close(){
err = client.Disconnect(context.TODO())

if err != nil {
    log.Fatal(err)
}
fmt.Println("Connection to MongoDB closed.")
}

func main() {
  Connect()
  time.Sleep(10 * time.Second)
  close()
}
