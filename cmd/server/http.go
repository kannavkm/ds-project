package main

import (
	"go.uber.org/zap"
	"net"
)

// http server for client requests

type httpServer struct {
	address net.Addr
	logger  *zap.Logger
}

func (server *httpServer) Start() {

}
