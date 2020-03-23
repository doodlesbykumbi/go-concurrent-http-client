package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)


func main() {
	var requestDuration int
	flag.IntVar(&requestDuration, "request-duration", 5, "request duration in seconds")
	flag.Parse()

	server := http.Server{
		Addr:      ":8080",
	}
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("Received request")
		time.Sleep(time.Duration(requestDuration)*time.Second)
		writer.WriteHeader(200)
	})

	log.Println("Running server at :8080")
	_ = server.ListenAndServe()
}
