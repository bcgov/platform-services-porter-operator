package portmap

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	minPort, _ = strconv.Atoi(os.Getenv("MIN_PORT"))
	maxPort, _ = strconv.Atoi(os.Getenv("MAX_PORT"))
)

type PortMap struct {
	portMap map[int]bool
	sync.Mutex
}

func Logger(logmsg string) {
    t := time.Now().Format(time.RFC3339)
    fmt.Println(t,": ",logmsg)
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
			logmsg := "* port: " + strconv.Itoa(port) + " - taken: " + strconv.FormatBool(taken)
			Logger(logmsg)
		}
	}
}

func (p *PortMap) ClaimHandler(w http.ResponseWriter, r *http.Request) {
	if port, err := p.ClaimUnusedPort(); err == nil {
		logmsg := "* port " + strconv.Itoa(port) + " claimed"
		Logger(logmsg)
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
		logmsg := "* port " + strconv.Itoa(port) + " relinquished"
		Logger(logmsg)
		json.NewEncoder(w).Encode(port)
	}
}

/*
Users may want to know ahead of time if a certain port is available.  Provide an
  API for this purpose.
A successful request returns the port number and a response code of 200 (OK).
If the port is not available, it returns a message with code 409 (Conflict).

Examples:
$ curl -w '%{response_code}\n' "http://localhost:9999/isPortAvailable?port=9999"
Port 9999 is not available
409

$ curl --fail-with-body "http://localhost:9999/isPortAvailable?port=9999"
Port 9999 is not available
curl: (22) The requested URL returned error: 409

$ curl -w '%{response_code}\n' "http://localhost:9999/isPortAvailable?port=9998"
Port 9998 is available
200
*/
func (p *PortMap) PortCheckHandler(w http.ResponseWriter, r *http.Request) {

	portRequest := r.URL.Query().Get("port")

	// Ensure that we got a port in the request
	if portRequest == "" {
		http.Error(w, "Port is missing from request.  Port should be in query string, such as: /isPortAvailable?port=12345", http.StatusBadRequest)
		return
	}

	// Make sure that it is an integer and in the acceptable range
	port, err := strconv.Atoi(portRequest)
	if err != nil || port < minPort || port > maxPort {
		errorMsg := fmt.Sprintf("Invalid port number: %d", port)
		http.Error(w, errorMsg, http.StatusBadRequest)
		return
	}

	// Report a conflict if the port is already in use
	if p.portMap[port] {
		errorMsg := fmt.Sprintf("Port %d is not available", port)
		http.Error(w, errorMsg, http.StatusConflict)
		return
	}

	// Otherwise, report that the requested port is available
	fmt.Fprintf(w, "Port %d is available\n", port)

}
