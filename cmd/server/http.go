package main

import (
	"go.uber.org/zap"
	"net"
)

type httpServer struct {
	address net.Addr
	logger  *zap.Logger
}

func (server *httpServer) Start() {

}
