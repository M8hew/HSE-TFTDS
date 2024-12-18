package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	handler "crdt/internal/handler"
)

type Config struct {
	ReplicaID string
	Port      int
	Peers     []string
}

func parseConfig() Config {
	var (
		replicaID string
		port      int
	)

	flag.StringVar(&replicaID, "id", "", "Id for replicaset")
	flag.IntVar(&port, "port", 0, "HTTP-port for replica")
	flag.Parse()

	return Config{
		ReplicaID: replicaID,
		Port:      port,
		Peers:     flag.Args(),
	}
}

func main() {
	cfg := parseConfig()

	svc := handler.NewReplica(cfg.ReplicaID, cfg.Peers)

	http.HandleFunc("/state", svc.HandleState)
	http.HandleFunc("/update", svc.HandleUpdate)
	http.HandleFunc("/switch", svc.HandleSwitch)
	http.HandleFunc("/sync", svc.HandleSync)

	fmt.Println("Starting server...")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))
}
