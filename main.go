package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

const (
	Attempts HealthCheck = iota
	Retry
)

const DEFAULT_PORT = 8000

var nodesPool NodesPool
var configs Configs

func main() {
	configsFilename := flag.String("configs", "configs.yaml", "the file containing the port and nodes list")
	flag.Parse()
	configs.load(*configsFilename)

	if len(configs.Nodes) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	nodesPool.RegisterNodes(configs.Nodes)

	server := http.Server{
		Addr:    fmt.Sprintf(":%s", configs.Port),
		Handler: http.HandlerFunc(lb),
	}

	go healthCheck()

	log.Printf("Load Balancer started at :%s\n", configs.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
