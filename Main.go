package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"time"
)

type userData struct {
	GUID    string
	Access  string
	Refresh []byte
}

var client *mongo.Client

func main() {
	fmt.Println("Application run")
	ConnectDB()
	handlers()
	closeDB()
}
func errExc(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func ConnectDB() {
	var err error
	client, err = mongo.NewClient(options.Client().ApplyURI(
		"mongodb+srv://user:user@clustervs.rwrgh.mongodb.net/" +
			"BaseOne?retryWrites=true&w=majority"))
	errExc(err)
	errExc(client.Connect(context.TODO()))
	errExc(client.Ping(context.TODO(), nil))
	fmt.Println("Connection to MongoDB")
}

func closeDB() {
	errExc(client.Disconnect(context.TODO()))
	fmt.Println("Connection to MongoDB closed")
}

func handlers() {
	router := mux.NewRouter()

	var receive = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("user")
		if username == "" {
			_, err := w.Write([]byte("Please write username: /receive?user=\"username\""))
			errExc(err)
			return
		}
		AccessToken, RefreshToken := receiver(username)
		collection := client.Database("BaseOne").Collection("ACol")
		session, err := client.StartSession()
		errExc(err)
		errExc(session.StartTransaction())
		bcryptHashResresh, err := bcrypt.GenerateFromPassword([]byte(RefreshToken), bcrypt.DefaultCost)
		_, err = collection.InsertOne(context.TODO(), userData{GUID: username, Access: AccessToken, Refresh: bcryptHashResresh})
		errExc(err)
		errExc(session.CommitTransaction(context.TODO()))
		session.EndSession(context.TODO())
		RefreshTokenBase64 := base64.StdEncoding.EncodeToString([]byte(RefreshToken))
		_, err = w.Write([]byte("username: \"" + username + "\" acquires a new tokens " +
			"\nAccess token: " + AccessToken +
			"\nRefresh token in base64: " + RefreshTokenBase64))
		errExc(err)
	})

	var refresh = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refresh, err := base64.StdEncoding.DecodeString(r.FormValue("refresh"))
		errExc(err)
		if len(refresh) < 2 {
			_, err := w.Write([]byte("Please write refresh token: /delete?refresh=\"refresh token\""))
			errExc(err)
			return
		}
		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(string(refresh), claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("WitmanStas"), nil
		})
		errExc(err)
		collection := client.Database("BaseOne").Collection("ACol")
		session, err := client.StartSession()
		errExc(err)
		errExc(session.StartTransaction())
		username := claims["user"].(string)
		res, err := collection.Find(context.TODO(), bson.D{{"guid", username}})
		errExc(err)
		flag := ""
		for res.Next(context.TODO()) {
			var users userData
			err = res.Decode(&users)
			errExc(err)
			err = bcrypt.CompareHashAndPassword(users.Refresh, refresh)
			if err == nil {
				AccessToken, RefreshToken := receiver(username)
				bcryptHashResresh, err := bcrypt.GenerateFromPassword([]byte(RefreshToken), bcrypt.DefaultCost)
				errExc(err)
				var users userData
				err = res.Decode(&users)
				result := collection.FindOneAndUpdate(context.TODO(), bson.D{{"refresh", users.Refresh}}, userData{GUID: username, Access: AccessToken, Refresh: bcryptHashResresh})
				fmt.Println(result)
				RefreshTokenBase64 := base64.StdEncoding.EncodeToString([]byte(RefreshToken))
				_, err = w.Write([]byte("Document of the user was updated" +
					"\nUsername: \"" + username + "\" acquires a new tokens " +
					"\nAccess token: " + AccessToken +
					"\nRefresh token in base64: " + RefreshTokenBase64))
				errExc(err)
				flag = "update"
			}
		}
		if flag == "" {
			_, err := w.Write([]byte("The refresh token was not found in DB"))
			errExc(err)
		} else {
			_, err = w.Write([]byte("Document of the user was updated"))
			errExc(err)
		}

		errExc(session.CommitTransaction(context.TODO()))
		session.EndSession(context.TODO())

	})
	var del = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refresh, err := base64.StdEncoding.DecodeString(r.FormValue("refresh"))
		errExc(err)
		if len(refresh) < 2 {
			_, err := w.Write([]byte("Please write refresh token: /delete?refresh=\"refresh token\""))
			errExc(err)
			return
		}
		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(string(refresh), claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("WitmanStas"), nil
		})
		errExc(err)

		collection := client.Database("BaseOne").Collection("ACol")
		session, err := client.StartSession()
		errExc(err)
		errExc(session.StartTransaction())
		username := claims["user"]
		res, err := collection.Find(context.TODO(), bson.D{{"guid", username}})
		errExc(err)
		flag := ""
		for res.Next(context.TODO()) {
			var users userData
			err = res.Decode(&users)
			errExc(err)
			err = bcrypt.CompareHashAndPassword(users.Refresh, refresh)
			if err == nil {
				_, err = collection.DeleteOne(context.TODO(), bson.D{{"refresh", users.Refresh}})
				errExc(err)
				flag = "delete"
			}
		}
		if flag == "" {
			_, err := w.Write([]byte("The refresh token was not found in DB"))
			errExc(err)
		} else {
			_, err = w.Write([]byte("Document of the user was deleted"))
			errExc(err)
		}

		errExc(session.CommitTransaction(context.TODO()))
		session.EndSession(context.TODO())

	})
	var clear = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("user")
		if username == "" {
			_, err := w.Write([]byte("Please write username: /clear?user=\"username\""))
			errExc(err)
			return
		}
		collection := client.Database("BaseOne").Collection("ACol")
		session, err := client.StartSession()
		errExc(err)
		errExc(session.StartTransaction())
		_, err = collection.DeleteMany(context.TODO(), bson.D{{"guid", username}})
		errExc(err)
		errExc(session.CommitTransaction(context.TODO()))
		session.EndSession(context.TODO())
		_, err = w.Write([]byte("Documents of the \"" + username + "\" user were deleted"))
		errExc(err)
	})

	router.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(" +/receive?user=...  = get the access and refresh tokens \n" +
			" +/refresh?refreshtoken=...  = refreshing access and refresh tokens \n" +
			" +/delete?refreshtoken=...  = delete the refresh token \n" +
			" +/clear?user=...  = delete all refresh tokens for one user \n"))
		errExc(err)
	})).Methods("Get")

	router.Handle("/receive", receive).Methods("Get")

	router.Handle("/refresh", refresh).Methods("Get")

	router.Handle("/delete", del).Methods("Get")

	router.Handle("/clear", clear).Methods("Get")

	port := os.Getenv("PORT")
	fmt.Println("Listen on port: " + port)
	errExc(http.ListenAndServe(":"+port, router))
}

func receiver(user string) (string, string) {
	var mySigningKey = []byte("WitmanStas")
	AccessToken := jwt.New(jwt.SigningMethodHS512)
	claimsAccess := make(jwt.MapClaims)
	claimsAccess["user"] = user
	claimsAccess["name"] = "AccessToken"
	claimsAccess["exp"] = time.Now().Add(time.Minute * 20).Unix()
	AccessToken.Claims = claimsAccess
	tokenAccessString, err := AccessToken.SignedString(mySigningKey)
	errExc(err)

	RefreshToken := jwt.New(jwt.SigningMethodHS256)
	claimsRefresh := make(jwt.MapClaims)
	claimsRefresh["user"] = user
	claimsRefresh["name"] = "RefreshToken"
	claimsRefresh["exp"] = time.Now().Add(time.Hour * 240).Unix()
	RefreshToken.Claims = claimsRefresh
	tokenRefreshString, err := RefreshToken.SignedString(mySigningKey)
	errExc(err)
	return tokenAccessString, tokenRefreshString
}
