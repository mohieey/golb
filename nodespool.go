package main

import (
	"log"
	"net/url"
	"sync/atomic"
)

type NodesPool struct {
	backendNodes []*BackendNode
	current      uint64
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
