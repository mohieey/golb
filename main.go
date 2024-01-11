package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const (
	Attempts = iota
	Retry
)

const DEFAULT_PORT = 8000

var nodesPool NodesPool

func main() {
	serversList := flag.String("backends", "", "Load balanced backends, use commas to separate")
	port := flag.Int("port", DEFAULT_PORT, "Port to serve")
	flag.Parse()

	if len(*serversList) == 0 {
		log.Fatal("Please provide one or more backends to load balance")
	}

	serversStrings := strings.Split(*serversList, ",")
	for _, serverString := range serversStrings {
		serverUrl, err := url.Parse(serverString)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(serverUrl)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			log.Printf("[%s] %s\n", serverUrl.Host, e.Error())

			retries := GetRetryFromContext(r)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(r.Context(), Retry, retries+1)
					proxy.ServeHTTP(w, r.WithContext(ctx))
				}
				return
			}

			nodesPool.MarkBackendStatus(serverUrl, false)

			attempts := GetAttemptsFromContext(r)
			log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
			ctx := context.WithValue(r.Context(), Attempts, attempts+1)
			fmt.Println("switch retrying")
			lb(w, r.WithContext(ctx))
		}

		nodesPool.AddBackend(&BackendNode{
			URL:          serverUrl,
			Alive:        true,
			ReverseProxy: proxy,
		})

		log.Printf("Configured server: %s\n", serverUrl)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: http.HandlerFunc(lb),
	}

	go healthCheck()

	log.Printf("Load Balancer started at :%d\n", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}
