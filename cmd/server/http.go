package main

import (
	"go.uber.org/zap"
	"net/http"
)

// methods
// 1. join (dynamically adding nodes should be a thing)
// 2. GET /key
// 3. PUT /key
// 4. DELETE /key

func (server *server) handleKeyGet(w http.ResponseWriter, r *http.Request) {
}

func (server *server) handleKeyPut(w http.ResponseWriter, r *http.Request) {
}

func (server *server) handleKeyDelete(w http.ResponseWriter, r *http.Request) {
}

func (server *server) Start() {
	server.logger.Info("Server Starting", zap.String("address", server.address.String()))

}
