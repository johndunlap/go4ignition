package main

import (
	_ "embed"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"server/sites"
	"strconv"
	"strings"
	"syscall"
)

func main() {

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)

	// Start the HTTP server in a separate goroutine
	go func() {
		http.HandleFunc("/", virtualHostHandler)

		// Start the server
		log.Printf("Starting server on port " + strconv.Itoa(sites.Port) + "\n")
		err := http.ListenAndServe(":"+strconv.Itoa(sites.Port), nil)

		if errors.Is(err, http.ErrServerClosed) {
			log.Printf("server closed\n")
		} else if err != nil {
			log.Printf("error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	<-sigc
	log.Println("\nReceived interrupt signal, shutting down...")
	os.Exit(0)
}

func virtualHostHandler(res http.ResponseWriter, req *http.Request) {
	host := req.Host
	colonIndex := strings.Index(host, ":")

	if colonIndex > 3 {
		host = host[0:colonIndex]
	}

	sites.Serve(host, res, req)
}
