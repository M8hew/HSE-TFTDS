package main

import (
	"flag"
	"fmt"
	"sync"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"kvstorage/api/oapi"
	"kvstorage/internal/raft"
	handler "kvstorage/internal/server"
	"kvstorage/internal/storage"
)

type Config struct {
	ID       int64
	RaftPort int
	HTTPPort int
	Peers    []string
}

func parseFlags() Config {
	var cfg Config

	flag.Int64Var(&cfg.ID, "id", 0, "ID for using in raft protocol")
	flag.IntVar(&cfg.RaftPort, "raft_port", 0, "Port to linsten on for Raft communication")
	flag.IntVar(&cfg.HTTPPort, "http_port", 0, "Port to linsten on for REST-endpoints")

	flag.Parse()

	cfg.Peers = flag.Args()

	return cfg
}

func toPeerAddrs(in []string) []raft.PeerAddr {
	addrs := make([]raft.PeerAddr, 0, len(in))
	for _, addr := range in {
		addrs = append(addrs, raft.PeerAddr(addr))
	}
	return addrs
}

func main() {
	cfg := parseFlags()

	fmt.Println(cfg)

	logger := zap.Must(zap.NewProduction())
	defer logger.Sync()

	localStorage := storage.NewLocalStorage()

	raftServer := raft.NewRaftServer(cfg.ID, toPeerAddrs(cfg.Peers), localStorage, logger)

	serverHandler := handler.NewHandler(raftServer, localStorage, logger)
	router := echo.New()
	oapi.RegisterHandlers(router, serverHandler)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		raftServer.Start(fmt.Sprintf("0.0.0.0:%d", cfg.RaftPort))
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		router.Start(fmt.Sprintf(":%d", cfg.HTTPPort))
	}()

	logger.Info("Server started and working")

	wg.Wait()

	logger.Info("Server gracefully stopped")
}
