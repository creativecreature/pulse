package main

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
)

// The RPCServer exposes the CodeHarvestApp methods to the RPC client.
type RPCServer struct {
	rcvr     *CodeHarvestApp
	listener net.Listener
}

func (s *RPCServer) start() error {
	err := rpc.Register(s.rcvr)
	if err != nil {
		return err
	}

	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	s.listener = listener
	return http.Serve(listener, nil)
}

func (s *RPCServer) stop() {
	s.listener.Close()
}
