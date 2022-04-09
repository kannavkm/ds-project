package main

import (
	// "flag"
	// "fmt"
	"log"
	"net/http"

	// "os"
	// "strings"
	badger "github.com/dgraph-io/badger/v3"
)

func say_hello(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("<h1>Hello</h1>"))
}

func main() {

	// id := flag.Int("id", 0, "Id of the node")
	// port := flag.String("p", "8001", "Port to listen on")
	port := ":8001"
	// cluster_s := flag.String("cluster", "", "Comma Separated ips of the rafts of current nodes")
	// var cluster []string = strings.Split(*cluster_s, ",")
	// for i := range cluster {
	// 	cluster[i] = strings.Trim(cluster[i], " ")
	// }
	// if len(*port) != 4 {
	// 	fmt.Println("Usage server [-p] port ...")
	// 	flag.PrintDefaults()
	// 	os.Exit(1)
	// }
	// flag.Parse()
	// fmt.Printf("Running Node: %d at port: %s", *id, *port)

	http.HandleFunc("/hello", say_hello)
	// go startBadger()

	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
			
			
			item.Value(func(val []byte) error {
				w.Write(val)
				return nil
			})


			
			return err
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	http.HandleFunc("/add", add)
	http.HandleFunc("/get", get)

	log.Fatal(http.ListenAndServe(port, nil))
}
