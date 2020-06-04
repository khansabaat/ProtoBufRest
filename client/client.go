package main

import (
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	pb "github.com/khansabaat/protofiles"
	"io/ioutil"
	"log"
	"net/http"
)

func makeRequest(request *pb.UserID) *pb.Retrieve {

	req, err := proto.Marshal(request)
	if err != nil {
		log.Fatalf("Unable to marshal request : %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(req)
	resp, err := http.Get(fmt.Sprintf("http://0.0.0.0:8000/user?proto_body=%v", encoded))

	if err != nil {
		log.Fatalf("Unable to read from the server : %v", err)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("Unable to read bytes from request : %v", err)
	}

	respObj := &pb.Retrieve{}
	proto.Unmarshal(respBytes, respObj)
	return respObj

}

func main() {

	payload := pb.UserID{Userid: "5ed935e124f4979db74c5f1d"}
	resp := makeRequest(&payload)
	fmt.Printf("Response from API is : %v\n", resp)
}