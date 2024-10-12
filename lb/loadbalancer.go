package lb

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/goyalanurag/go-load-balancer/peer"
)

type LoadBalancer struct {
	peerGroup         []*peer.Peer
	lastUsedPeerIndex int
}

// https://stackoverflow.com/questions/36773837
func (lb *LoadBalancer) peerHealthMonitor(interval time.Duration) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	healthMonitorRoutineGroup := sync.WaitGroup{}

	for _, peerServer := range lb.peerGroup {
		healthMonitorRoutineGroup.Add(1)

		go func() {
			defer healthMonitorRoutineGroup.Done()

			for {
				err := peerServer.HealthCheck(client)
				if err != nil {
					log.Printf("Healthcheck %s: DOWN", peerServer.URL.String())
				} else {
					log.Printf("Healthcheck %s: UP", peerServer.URL.String())
				}

				time.Sleep(interval)
			}
		}()
	}

	healthMonitorRoutineGroup.Wait()
}

// round robin
func (lb *LoadBalancer) getNextAvailablePeer() *peer.Peer {
	peerCount := len(lb.peerGroup)

	for i := range peerCount {
		newPeerIndex := (lb.lastUsedPeerIndex + i + 1) % peerCount
		if lb.peerGroup[newPeerIndex].IsAlive() {
			lb.lastUsedPeerIndex = newPeerIndex
			return lb.peerGroup[newPeerIndex]
		}
	}

	return nil
}

func (lb *LoadBalancer) Init(peerServers []string) error {
	for _, peerServer := range peerServers {
		peerURL, err := url.Parse(peerServer)
		if err != nil {
			lb.peerGroup = nil
			log.Fatal(err)
			return err
		}

		lb.peerGroup = append(lb.peerGroup, &peer.Peer{
			URL:          peerURL,
			ReverseProxy: httputil.NewSingleHostReverseProxy(peerURL),
		})
	}

	go lb.peerHealthMonitor(1 * time.Minute)

	return nil
}

// https://stackoverflow.com/questions/17758345
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer := lb.getNextAvailablePeer()
	if peer != nil {
		peer.ReverseProxy.ServeHTTP(w, r)
		return
	}

	w.WriteHeader(http.StatusServiceUnavailable)
	fmt.Fprintf(w, "Service Unavailable")
}
