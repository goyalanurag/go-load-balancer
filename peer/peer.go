package peer

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Peer struct {
	mux          sync.Mutex
	URL          *url.URL
	alive        bool
	ReverseProxy *httputil.ReverseProxy
}

func (p *Peer) IsAlive() bool {
	p.mux.Lock()
	status := p.alive
	p.mux.Unlock()

	return status
}

func (p *Peer) setAlive(status bool) {
	p.mux.Lock()
	p.alive = status
	p.mux.Unlock()
}

func (p *Peer) HealthCheck(client *http.Client) error {
	resp, err := client.Get(p.URL.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		p.setAlive(false)
		return err
	}

	defer resp.Body.Close()
	p.setAlive(true)

	return nil
}
