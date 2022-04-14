package main

import (
	"go.uber.org/zap"
	"net"
	"net/http"
)

// http server for client requests
type httpServer struct {
	address net.Addr
	logger  *zap.Logger
}

func (server *httpServer) Start() {
	server.logger.Info("Server Starting", zap.String("address", server.address.String()))
	if err := http.ListenAndServe(server.address.String(), )
}
