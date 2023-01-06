package main

import (
	"fmt"
	"net/http"
	"flag"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sqs"
)

func VerifyRouteAndMethod(w http.ResponseWriter, req *http.Request, route string, method string) bool {
	if req.URL.Path != route || req.Method != method {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return false
	}
	return true
}

func GetItem(w http.ResponseWriter, req *http.Request) {
	if !VerifyRouteAndMethod(w, req, "/getItem", "GET") {
		return
	}

	fmt.Fprintf(w, "Hello, %s!", req.URL.Path[1:])
}

func GetItems(w http.ResponseWriter, req *http.Request) {
	if !VerifyRouteAndMethod(w, req, "/getItems", "GET") {
		return
	}

	fmt.Fprintf(w, "Hello, %s!", req.URL.Path[1:])
}

func AddItem(w http.ResponseWriter, req *http.Request) {
	if !VerifyRouteAndMethod(w, req, "/addItem", "POST") {
		return
	}

	fmt.Fprintf(w, "Hello, %s!", req.URL.Path[1:])
}

func RemoveItem(w http.ResponseWriter, req *http.Request) {
	if !VerifyRouteAndMethod(w, req, "/removeItem", "DELETE") {
		return
	}

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {
	http.HandleFunc("/getItem", GetItem)
	http.HandleFunc("/getItems", GetItems)
	http.HandleFunc("/addItem", AddItem)
	http.HandleFunc("/removeItem", RemoveItem)
	http.ListenAndServe(":8080", nil)
}
