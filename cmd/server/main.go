package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func say_hello(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("<h1>Hello</h1>"))
}

func main() {
	id := flag.Int("id", 0, "Id of the node")
	port := flag.String("p", "8001", "Port to listen on")
	cluster_s := flag.String("cluster", "", "Comma Separated ips of the rafts of current nodes")
	var cluster []string = strings.Split(*cluster_s, ",")
	for i := range cluster {
		cluster[i] = strings.Trim(cluster[i], " ")
	}
	if len(*port) != 4 {
		fmt.Println("Usage server [-p] port ...")
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	fmt.Printf("Running Node: %d at port: %s", *id, *port)
	http.HandleFunc("/hello", say_hello)
	log.Fatal(http.ListenAndServe(cluster[*id], nil))
}
