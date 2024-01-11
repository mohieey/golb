package main

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type BackendNode struct {
	URL          *url.URL
	Alive        bool
	mux          sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (bn *BackendNode) SetAlive(alive bool) {
	bn.mux.Lock()
	bn.Alive = alive
	bn.mux.Unlock()
}

func (bn *BackendNode) IsAlive() bool {
	bn.mux.RLock()
	alive := bn.Alive
	bn.mux.RUnlock()
	return alive
}
