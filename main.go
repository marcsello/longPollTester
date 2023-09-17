package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	desiredTimeout time.Duration
	nextId         atomic.Int64
)

type Handler struct {
}

func (_ Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	connectionStarted := time.Now()
	var ctx *context.Context
	if desiredTimeout != 0 {
		c, cancel := context.WithTimeout(r.Context(), desiredTimeout)
		defer cancel()
		ctx = &c
	} else {
		c := r.Context()
		ctx = &c
	}

	reqId := nextId.Add(1)
	log.Printf("[%d] New %s request from %s to %s\n", reqId, r.Method, r.RemoteAddr, r.URL)

	<-(*ctx).Done()
	reason := (*ctx).Err()

	log.Printf("[%d] Connection closed! Reason: %s. Was open for %s", reqId, reason, time.Since(connectionStarted))
}

func main() {
	bindAddr := flag.String("bind", ":8080", "Address string to bind the server to")
	flag.DurationVar(&desiredTimeout, "timeout", 0, "Time for the server to return with 204 after, don't set it to never return (for intermediate timeout tests)")
	flag.Parse()

	log.Printf("Starting HTTP server on %s\n", *bindAddr)
	err := http.ListenAndServe(*bindAddr, Handler{})
	if err != nil {
		log.Fatal(err)
	}
}
