package main

import (
	"flag"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"go.uber.org/zap"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	logger.Info("Hello from zap logger")

	id := flag.Int("id", 0, "Id of the cluster")
	port := flag.String("p", "8001", "Port to listen on")
	clusterS := flag.String("cluster", "", "Comma Separated ips of the rafts of current nodes")

	if len(*port) != 4 {
		fmt.Println("Usage server [-p] port ...")
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	var cluster []string = strings.Split(*clusterS, ",")
	for i := range cluster {
		cluster[i] = strings.Trim(cluster[i], " ")
	}
	fmt.Printf("Running Node: %d at port: %s", *id, *port)
	fmt.Println(cluster)
	// go startBadger()

	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *badger.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(db)

	add := func(w http.ResponseWriter, r *http.Request) {
		keys := r.URL.Query()["key"]
		vals := r.URL.Query()["val"]

		key := keys[0]
		val := vals[0]

		err := db.Update(func(txn *badger.Txn) error {
			err := txn.Set([]byte(key), []byte(val))
			return err
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	get := func(w http.ResponseWriter, r *http.Request) {
		keys := r.URL.Query()["key"]
		key := keys[0]

		err := db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(key))
			if err != nil {
				return err
			}
			err = item.Value(func(val []byte) error {
				_, err := w.Write(val)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}

			return err
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	sayHello := func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write([]byte("<h1>Hello</h1>"))
		if err != nil {
			return
		}
	}

	http.HandleFunc("/", sayHello)
	http.HandleFunc("/add", add)
	http.HandleFunc("/get", get)

	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
