package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type httpService struct {
	addr   string
	logger *zap.Logger
	server *ZeroServer
	c      *consistentHashHandler
}

const (
	SEPARATOR = "%"
)

func (s *httpService) handleKeyOps(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	relation := vars["relation"]
	predicate := key + SEPARATOR + relation
	mem := s.c.c.LocateKey([]byte(predicate))
	grp, err := s.server.GetGroupInfo(mem.String())
	if err != nil {
		http.Error(w, "Could not get the keyinfo", 500)
		return
	}
	type response struct {
		Id    string `json:"id"`
		HAddr string `json:"haddr"`
		RAddr string `json:"raddr"`
		Mem   int    `json:"mem"`
	}
	resp := response{
		Id:    grp.GetId(),
		HAddr: grp.GetLeaderHttpAddress(),
		RAddr: grp.GetLeaderRaftAddress(),
		Mem:   int(grp.GetMembers()),
	}
	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Could marshal data", 500)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		http.Error(w, "Error in writing response", 500)
		return
	}
}

func (s *httpService) Start() {
	s.logger.Info("Server Starting", zap.String("address", s.addr))
	r := mux.NewRouter()
	r.HandleFunc("/{id}/{relation}", s.handleKeyOps).Methods("GET", "PUT")
	http.Handle("/", r)
	srv := http.Server{
		Handler: r,
		Addr:    s.addr,
	}
	log.Fatal(srv.ListenAndServe())
}
