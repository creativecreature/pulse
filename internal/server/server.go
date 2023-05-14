package server

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"

	"code-harvest.conner.dev/internal/domain"
	"code-harvest.conner.dev/internal/proxy"
	"code-harvest.conner.dev/internal/storage"
)

type server struct {
	serverName     string
	mutex          sync.Mutex
	clock          Clock
	metadataReader MetadataReader
	storage        storage.Storage
	activeClientId string
	lastHeartbeat  int64
	session        *domain.Session
	log            Log
}

func (server *server) Start(port string) error {
	server.log.PrintInfo("Starting up...", nil)

	// Connect to the storage
	disconnect := server.storage.Connect()
	defer disconnect()

	// Start the RPC server
	listener, err := startServer(server, port)
	if err != nil {
		server.log.PrintFatal(err, nil)
	}

	// Blocks until we receive a shutdown signal
	server.monitorHeartbeat()

	server.log.PrintInfo("Shutting down...", nil)
	return listener.Close()
}

func startServer(server *server, port string) (net.Listener, error) {
	// The proxy exposes the functions that we want to make available for remote
	// procedure calls. Register the proxy as the RPC receiver.
	proxy := proxy.New(server)
	err := rpc.RegisterName(server.serverName, proxy)
	if err != nil {
		return nil, err
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return nil, err
	}

	err = http.Serve(listener, nil)
	return listener, err
}
