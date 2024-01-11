package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"
)

type NodesPool struct {
	backendNodes []*BackendNode
	current      uint64
}

func (np *NodesPool) RegisterNodes(nodesURLs []string) {
	for _, nodeURL := range nodesURLs {
		nodeURL, err := url.Parse(nodeURL)
		if err != nil {
			log.Fatal(err)
		}

		proxy := httputil.NewSingleHostReverseProxy(nodeURL)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
			log.Printf("[%s] %s\n", nodeURL.Host, e.Error())

			retries := GetRetryFromContext(r)
			if retries < 3 {
				select {
				case <-time.After(10 * time.Millisecond):
					ctx := context.WithValue(r.Context(), Retry, retries+1)
					proxy.ServeHTTP(w, r.WithContext(ctx))
				}
				return
			}

			nodesPool.MarkBackendStatus(nodeURL, false)

			attempts := GetAttemptsFromContext(r)
			log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
			ctx := context.WithValue(r.Context(), Attempts, attempts+1)
			lb(w, r.WithContext(ctx))
		}

		nodesPool.AddBackend(&BackendNode{
			URL:          nodeURL,
			Alive:        true,
			ReverseProxy: proxy,
		})

		log.Printf("Configured server: %s\n", nodeURL)
	}
}

func (np *NodesPool) AddBackend(backend *BackendNode) {
	np.backendNodes = append(np.backendNodes, backend)
}

func (np *NodesPool) MarkBackendStatus(backendUrl *url.URL, alive bool) {
	for _, b := range np.backendNodes {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(alive)
			break
		}
	}
}

func (np *NodesPool) NextIndex() int {
	return int(atomic.AddUint64(&np.current, uint64(1)) % uint64(len(np.backendNodes)))
}

func (np *NodesPool) GetNextPeer() *BackendNode {
	next := np.NextIndex()
	l := len(np.backendNodes) + next

	for i := next; i < l; i++ {
		idx := i % len(np.backendNodes)

		if np.backendNodes[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&np.current, uint64(idx))
			}

			return np.backendNodes[idx]
		}
	}

	return nil
}

func (np *NodesPool) HealthCheck() {
	for _, b := range np.backendNodes {
		status := "up"
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			status = "down"
		}
		log.Printf("%s [%s]\n", b.URL, status)
	}
}
