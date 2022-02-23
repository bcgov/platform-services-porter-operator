package portmap

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type PortMap struct {
	portMap map[int]bool
	sync.Mutex
}

func NewPortMap(min int, max int) *PortMap {
	tempPortMap := make(map[int]bool)
	for i := min; i <= max; i++ {
		tempPortMap[i] = false
	}

	return &PortMap{
		portMap: tempPortMap,
	}
}

func (p *PortMap) ClaimUnusedPort() (int, error) {
	p.Lock()
	defer p.Unlock()

	for port, taken := range p.portMap {
		if !taken {
			p.portMap[port] = true
			return port, nil
		}
	}
	return -1, fmt.Errorf("out of ports or something has gone terribly wrong")
}

func (p *PortMap) ClaimPort(port int) bool {
	p.Lock()
	defer p.Unlock()
	if p.portMap[port] {
		return false
	}
	p.portMap[port] = true
	return true
}

func (p *PortMap) RelinquishPort(port int) bool {
	p.Lock()
	defer p.Unlock()
	if p.portMap[port] {
		p.portMap[port] = false
		return true
	}
	return false
}

func (p *PortMap) PrintClaimedPorts() {
	for port, taken := range p.portMap {
		if taken {
			fmt.Printf("* port: %d - taken: %t\n", port, taken)
		}
	}
}

func (p *PortMap) ClaimHandler(w http.ResponseWriter, r *http.Request) {
	if port, err := p.ClaimUnusedPort(); err == nil {
		fmt.Printf("* port %d claimed\n", port)
		json.NewEncoder(w).Encode(port)
	} else {
		json.NewEncoder(w).Encode(err)
		log.Fatal(err)
	}
}

func (p *PortMap) RelinquishHandler(w http.ResponseWriter, r *http.Request) {
	port, err := strconv.Atoi(r.FormValue("port"))
	if err != nil {
		fmt.Println(err)
	}

	if p.portMap[port] {
		p.portMap[port] = false
		fmt.Printf("* port %d relinquished\n", port)
		json.NewEncoder(w).Encode(port)
	}
}
