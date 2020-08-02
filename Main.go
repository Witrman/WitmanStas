package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
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
var mySigningKey = []byte("WitmanStas")
var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	},
	SigningMethod: jwt.SigningMethodHS512,
})

func main() {
	fmt.Println("Server run")
	ConnectDB()
	handlers()
	closeDB()
}
func errExc(err error) {
	if err != nil {
		fmt.Println(err)
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
	router.Handle("/receive", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("user")
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
		_, err = w.Write([]byte("Access: " + AccessToken +
			" Refresh: " + RefreshTokenBase64 + " User: " + username))
		errExc(err)
		fmt.Println("Get tokens")

	})).Methods("Get")

	router.Handle("/refresh", jwtMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refresh, err := base64.StdEncoding.DecodeString(r.Header.Get("refresh"))
		if err != nil {
			_, err := w.Write([]byte("Некорректный токен обновления"))
			errExc(err)
			fmt.Println("Invalid refresh token")
			return
		}
		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(string(refresh), claims, func(token *jwt.Token) (interface{}, error) {
			return mySigningKey, nil
		})
		if err != nil {
			_, err := w.Write([]byte("Некорректный токен обновления"))
			errExc(err)
			fmt.Println("Invalid refresh token")
			return
		}
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
			errExc(res.Decode(&users))
			err = bcrypt.CompareHashAndPassword(users.Refresh, refresh)
			if err == nil {
				AccessToken, RefreshToken := receiver(username)
				bcryptHashResresh, err := bcrypt.GenerateFromPassword([]byte(RefreshToken), bcrypt.DefaultCost)
				errExc(err)
				var users userData
				errExc(res.Decode(&users))
				result := collection.FindOneAndReplace(context.TODO(), bson.D{{"refresh", users.Refresh}}, userData{GUID: username, Access: AccessToken, Refresh: bcryptHashResresh})
				errExc(result.Err())
				RefreshTokenBase64 := base64.StdEncoding.EncodeToString([]byte(RefreshToken))
				_, err = w.Write([]byte("Access: " + AccessToken +
					" Refresh: " + RefreshTokenBase64))
				errExc(err)
				flag = "update"
				fmt.Println("Tokens were refreshed")
			}
		}
		if flag == "" {
			_, err := w.Write([]byte("Токен обновления не был найден в базе данных"))
			errExc(err)
			fmt.Println("Refresh token was not found in DB")
		}
		errExc(session.CommitTransaction(context.TODO()))
		session.EndSession(context.TODO())

	}))).Methods("Get")

	router.Handle("/delete", jwtMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refresh, err := base64.StdEncoding.DecodeString(r.Header.Get("refresh"))
		if err != nil {
			_, err := w.Write([]byte("Некорректный токен обновления"))
			errExc(err)
			fmt.Println("Invalid refresh token")
			return
		}
		claims := jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(string(refresh), claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("WitmanStas"), nil
		})
		if err != nil {
			_, err := w.Write([]byte("Некорректный токен обновления"))
			errExc(err)
			fmt.Println("Invalid refresh token")
			return
		}
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
			errExc(res.Decode(&users))
			err = bcrypt.CompareHashAndPassword(users.Refresh, refresh)
			if err == nil {
				_, err = collection.DeleteOne(context.TODO(), bson.D{{"refresh", users.Refresh}})
				errExc(err)
				flag = "delete"
				fmt.Println("Token was deleted")
			}
		}
		if flag == "" {
			_, err := w.Write([]byte("Токен обновления не был найден"))
			errExc(err)
			fmt.Println("Refresh token was not found in DB")
		} else {
			_, err = w.Write([]byte("Запись пользователя \"" + username + "\" была удалена"))
			errExc(err)
		}

		errExc(session.CommitTransaction(context.TODO()))
		session.EndSession(context.TODO())

	}))).Methods("Get")

	router.Handle("/clear", jwtMiddleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("user")
		collection := client.Database("BaseOne").Collection("ACol")
		session, err := client.StartSession()
		errExc(err)
		errExc(session.StartTransaction())
		deleteResult, err := collection.DeleteMany(context.TODO(), bson.D{{"guid", username}})
		errExc(err)
		str := "Записи пользователя \"" + username + "\" не были найдены в базе данных "
		if deleteResult.DeletedCount > 0 {
			str = "Записи пользователя \"" + username + "\" были удалены"
			fmt.Println("Documents were deleted")
		} else {
			fmt.Println("Documents were not found in DB")
		}
		errExc(session.CommitTransaction(context.TODO()))
		session.EndSession(context.TODO())
		_, err = w.Write([]byte(str))
		errExc(err)

	}))).Methods("Get")

	port := os.Getenv("PORT")
	fmt.Println("Listen on port: " + port)
	errExc(http.ListenAndServe(":"+port, router))
}

func receiver(user string) (string, string) {
	AccessToken := jwt.New(jwt.SigningMethodHS512)
	claimsAccess := make(jwt.MapClaims)
	claimsAccess["user"] = user
	claimsAccess["name"] = "AccessToken"
	claimsAccess["exp"] = time.Now().Add(time.Second * 20).Unix()
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
