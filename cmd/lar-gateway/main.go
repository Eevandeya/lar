package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eevandeya/lar/internal/api"
	"github.com/eevandeya/lar/internal/config"
	"github.com/eevandeya/lar/internal/gateway/arp"
)

const ServerPort = 8080

var cfg *config.GatewayConfig

func checkAuth(r *http.Request) bool {
	authHeader, ok := r.Header["Authorization"]
	if !ok || len(authHeader) == 0 {
		return false
	}

	// NOTE: simplification for now
	authPayload := authHeader[0]

	if !strings.HasPrefix(authPayload, "Bearer ") {
		return false
	}

	return strings.TrimPrefix(authPayload, "Bearer ") == cfg.Secret
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting to process status...")

	if !checkAuth(r) {
		log.Println("403")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	queryParams := r.URL.Query()
	hostName := queryParams.Get("host")

	if hostName == "" {
		log.Println("400: bad host name")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	host, ok := cfg.Hosts[hostName]
	if !ok {
		log.Println("400: no such host")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	iface, err := net.InterfaceByName(cfg.InterfaceName)
	if err != nil {
		log.Println("500: no interface")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	online, err := arp.Probe(net.HardwareAddr(host.MAC), net.IP(host.Address), iface)
	if err != nil {
		log.Printf("500: cant arp: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := api.StatusResponse{Online: online}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	fmt.Println("200: all good")

	if err != nil {
		log.Printf("Somehow this is where we are rn: %v\n", err)
		return
	}

	return
}

func main() {
	configPath := flag.String("config", "/etc/lar/config.yaml", "path to config file")
	flag.Parse()

	var err error
	cfg, err = config.LoadGateway(*configPath)
	if err != nil {
		log.Fatalf("Can not load config: %v\n", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /status", statusHandler)

	s := &http.Server{
		Addr:           ":" + strconv.Itoa(ServerPort),
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Starting server")

	log.Fatal(s.ListenAndServe())
}
