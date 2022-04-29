package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type httpService struct {
	addr   string
	store  *server
	logger *zap.Logger
}

func (s *httpService) handleKeyGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Wrong key", 400)
		return
	}
	value, err := s.store.get(key)
	if err != nil {
		http.Error(w, "Could not get the key", 500)
		return
	}
	valueS := strconv.FormatUint(value, 10)
	_, err = w.Write([]byte(valueS))
	if err != nil {
		http.Error(w, "Error in writing response", 500)
		return
	}
}

func (s *httpService) handleKeyPut(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Wrong key", 400)
		return
	}
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, "Could not open request body", 500)
		return
	}
	type message struct {
		value uint64
	}
	var msg message
	err = json.Unmarshal(b, &msg)
	if err != nil {
		http.Error(w, "Could not parse Request body", 400)
		return
	}
	value := msg.value
	err = s.store.put(key, value)
	if err != nil {
		http.Error(w, "Could not put the key", 500)
		return
	}
	valueS := strconv.FormatUint(value, 10)
	_, err = w.Write([]byte(valueS))
	if err != nil {
		http.Error(w, "Error in writing response", 500)
		return
	}
}

func (s *httpService) handleKeyDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Wrong key", 400)
		return
	}
	err = s.store.delete(key)
	if err != nil {
		http.Error(w, "Could not delete the key", 500)
		return
	}
	_, err = w.Write([]byte("0"))
	if err != nil {
		http.Error(w, "Error in writing response", 500)
		return
	}
}

func (s *httpService) handleJoin(w http.ResponseWriter, r *http.Request) {

}

func (s *httpService) Start() {
	s.logger.Info("Server Starting", zap.String("address", s.addr))
	r := mux.NewRouter()
	r.HandleFunc("/join", s.handleJoin)
	r.HandleFunc("/{id:[0-9]+}", s.handleKeyGet).Methods("GET")
	r.HandleFunc("/{id:[0-9]+}", s.handleKeyPut).Methods("PUT")
	r.HandleFunc("/{id:[0-9]+}", s.handleKeyDelete).Methods("DELETE")
	http.Handle("/", r)
	srv := http.Server{
		Handler: r,
		Addr:    s.addr,
	}
	log.Fatal(srv.ListenAndServe())
}
