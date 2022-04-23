package main

import (
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"log"
	"net"
	"net/http"
	"strconv"
)

type kvStore interface {
	// 1. join (dynamically adding nodes should be a thing)
	join(id uint64, addr net.Addr) error
	// 2. GET /key
	get(key uint64) (uint64, error)
	// 3. SET /key
	put(key, val uint64) error
	// 4. DELETE /key
	delete(key uint64) error
}

type httpService struct {
	addr   net.Addr
	srv    *http.Server
	store  kvStore
	logger *zap.Logger
}

func (s *httpService) handleKeyGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {

	}
	value, err := s.store.get(key)
	if err != nil {

	}
}

func (s *httpService) handleKeyPut(w http.ResponseWriter, r *http.Request) {

}

func (s *httpService) handleKeyDelete(w http.ResponseWriter, r *http.Request) {
}

func (s *httpService) handleJoin(w http.ResponseWriter, r *http.Request) {

}

func (s *httpService) Start() {
	s.logger.Info("Server Starting", zap.String("address", s.addr.String()))
	r := mux.NewRouter()
	r.HandleFunc("/join", s.handleJoin)
	r.HandleFunc("/{id:[0-9]+}", s.handleKeyGet).Methods("GET")
	r.HandleFunc("/{id:[0-9]+}", s.handleKeyPut).Methods("PUT")
	r.HandleFunc("/{id:[0-9]+}", s.handleKeyDelete).Methods("DELETE")
	http.Handle("/", r)
	srv := http.Server{
		Handler: r,
		Addr:    s.addr.String(),
	}
	log.Fatal(srv.ListenAndServe())
}
