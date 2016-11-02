package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

func main() {
	port := flag.Int("port", 3000, "port to serve on, e.g. 8080")
	servePath := flag.String("serve", "./static", "path to serve from, default: ./static")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// TBD: We could  add some JWT check on all requests
	r := mux.NewRouter()

	r.HandleFunc("/go-webserial.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./go-webserial.js")
	})

	// API calls
	r.HandleFunc("/commports", handleListCommPortsEvent).
		Methods("GET") // Change to POST later

	http.HandleFunc("/wsconnect", handleWSConnect)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(*servePath + "/")))

	http.Handle("/", r)

	fmt.Printf("The serial proxy service is serving on port %v\n", *port)
	http.ListenAndServe(":"+strconv.FormatInt(int64(*port), 10), nil)
}
