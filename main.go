package main

import (
	"log"
	"net/http"

	"github.com/goyalanurag/go-load-balancer/lb"
)

var peerServers = []string{
	"http://localhost:3001",
	"http://localhost:3002",
	"http://localhost:3003",
}

func main() {
	log.Println("Starting load balancer ...")

	lb := lb.LoadBalancer{}
	lb.Init(peerServers)

	http.HandleFunc("/", lb.ServeHTTP)
	err := http.ListenAndServe("127.0.0.1:3000", nil)

	if err != nil {
		log.Fatal(err)
	}
}
