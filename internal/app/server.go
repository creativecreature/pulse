package app

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"

	"code-harvest.conner.dev/internal/shared"
)

func startServer(app *app, port string) (net.Listener, error) {
	// The proxy exposes the functions that we want to make available for remote
	// procedure calls. Register the proxy as the RPC receiver.
	proxy := shared.NewProxy(app)
	err := rpc.RegisterName(shared.ServerName, proxy)
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
