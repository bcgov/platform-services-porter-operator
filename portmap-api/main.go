package main

import (
	pm "bcgov/portmap-api/portmap"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Healthy!")
}

func main() {
	var logmsg string

	//fmt.Println("Porter Sidecar Starting Up...")
	pm.Logger("Porter Sidecar Starting Up...")

	minPort, err := strconv.Atoi(os.Getenv("MIN_PORT"))
	if err != nil {
		log.Fatal("MIN_PORT is not a positive integer:", err)
	}

	maxPort, err := strconv.Atoi(os.Getenv("MAX_PORT"))
	if err != nil {
		log.Fatal("MAX_PORT is not a positive integer:", err)
	}

	fmt.Printf("MIN_PORT: %d - MAX_PORT: %d\n", minPort, maxPort)
	portMap := pm.NewPortMap(minPort, maxPort)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	transportServerRes := schema.GroupVersionResource{Group: "cis.f5.com", Version: "v1", Resource: "transportservers"}
	list, err := client.Resource(transportServerRes).Namespace("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	pm.Logger("Adding Existing TransportServer VirtualServerPorts to Claimed Ports...")
	for _, ts := range list.Items {
		virtualServerPort, found, err := unstructured.NestedInt64(ts.Object, "spec", "virtualServerPort")
		if err != nil || !found {
			logmsg = "virtualServerPort not found for TransportServer " + ts.GetName() + ": error=" + err.Error()
			pm.Logger(logmsg)
			continue
		}

		if portMap.ClaimPort(int(virtualServerPort)) {
			logmsg = " " + strconv.FormatInt(virtualServerPort, 10) + " claimed by: " + ts.GetName()
			pm.Logger(logmsg)
			continue
		}

		logmsg = "unable to claim virtualServerPort: " + strconv.FormatInt(virtualServerPort, 10)
		pm.Logger(logmsg)
	}

	http.HandleFunc("/", healthz)
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/claim", portMap.ClaimHandler)
	http.HandleFunc("/relinquish", portMap.RelinquishHandler)
	http.HandleFunc("/isPortAvailable", portMap.PortCheckHandler)

	log.Fatal(http.ListenAndServe(":10000", nil))
}
