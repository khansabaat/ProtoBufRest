package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	pb "github.com/khansabaat/protofiles"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var client *mongo.Client

func CreateUser(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	// Creating error message object
	failureMessage := pb.Failure{Details: "Error Occured."}
	fail, _ := proto.Marshal(&failureMessage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(fail)
	}
	// Reading request data
	requestData := &pb.Payload{}
	proto.Unmarshal(data, requestData)

	// Creating user object from request data
	user := pb.User{
		FirstName: requestData.FirstName,
		LastName:  requestData.LastName,
		Email:     requestData.Email,
	}
	// Inserting user into DB
	userCollection := client.Database("mfmdb").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := userCollection.InsertOne(ctx, &user)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(fail)
	}
	userId, _ := result.InsertedID.(primitive.ObjectID)
	employeeCollection := client.Database("mfmdb").Collection("employees")
	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.M{"designation": requestData.Designation}
	update := bson.M{"$push": bson.M{"userid": userId.Hex()}, "$set": bson.M{"designation": requestData.Designation}}
	var updatedDocument bson.M
	err = employeeCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&updatedDocument)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(fail)
	}
	// Creating Success message
	succesMessage := pb.Success{
		Details:  "Succesfully Created",
		ObjectId: userId.Hex(),
	}
	res, _ := proto.Marshal(&succesMessage)
	w.Write(res)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	failureMessage := pb.Failure{Details: "Error Occured."}
	fail, _ := proto.Marshal(&failureMessage)
	queryParams := r.URL.Query()
	protoBody, err := base64.StdEncoding.DecodeString(queryParams["proto_body"][0])
	userId := pb.UserID{}
	err = proto.Unmarshal(protoBody, &userId)
	id, _ := primitive.ObjectIDFromHex(userId.Userid)
	user := &pb.User{}
	collection := client.Database("mfmdb").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err = collection.FindOne(ctx, bson.M{"_id": id}).Decode(user)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(fail)
	}
	// Fetch employee object
	collection = client.Database("mfmdb").Collection("employees")
	var employee bson.M
	err = collection.FindOne(ctx, bson.M{"userid": bson.M{"$in": bson.A{userId.Userid}}}).Decode(&employee)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(fail)
	}
	response := pb.Retrieve{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		EmployeeId:  employee["_id"].(primitive.ObjectID).Hex(),
		Designation: employee["designation"].(string),
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(fail)
	}
	res, _ := proto.Marshal(&response)
	w.Write(res)
}

func UpdateUser(w http.ResponseWriter, r *http.Request){
	data, err := ioutil.ReadAll(r.Body)
	// Creating error message object
	failureMessage := pb.Failure{Details: "Error Occured."}
	fail, _ := proto.Marshal(&failureMessage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(fail)
	}
	// Reading request data
	user := &pb.UpdateUser{}
	proto.Unmarshal(data, user)

	// Inserting user into DB
	userCollection := client.Database("mfmdb").Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	id, _ := primitive.ObjectIDFromHex(user.Id)
	result, err := userCollection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.D{
			{"$set", bson.D{{"Email", user.Email}}},
		},
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(fail)
	}
	fmt.Printf("Updated %v Documents!\n", result.ModifiedCount)
	// Creating Success message
	succesMessage := pb.Success{
		Details:  "Succesfully Updated",
	}
	res, _ := proto.Marshal(&succesMessage)
	w.Write(res)
}

func main() {
	r := mux.NewRouter()
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	r.HandleFunc("/user", CreateUser).Methods("POST")
	r.HandleFunc("/user", GetUser).Methods("GET")
	r.HandleFunc("/user", UpdateUser).Methods("PATCH")
	log.Fatal(http.ListenAndServe(":8000", r))
}
