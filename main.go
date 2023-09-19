package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	nextId atomic.Int64

	// config (these are read-only, so we are safe)
	serverTimeout     time.Duration
	writeHeaderEarly  bool
	returnStatus      int
	payload           string
	keepAliveInterval time.Duration
	keepAlivePayload  string
)

type Handler struct {
}

func (_ Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	connectionStarted := time.Now()

	reqId := nextId.Add(1)
	log.Printf("[%d] New %s request from %s to %s\n", reqId, r.Method, r.RemoteAddr, r.URL)

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Fatal("Could not acquire ResponseWriter flusher")
	}

	if writeHeaderEarly {
		w.WriteHeader(returnStatus)
		flusher.Flush()
	}

	var timeoutChan <-chan time.Time
	if serverTimeout != 0 {
		timeoutChan = time.After(serverTimeout)
	}

	var keepaliveTickerChan <-chan time.Time
	if keepAliveInterval != 0 {
		t := time.NewTicker(keepAliveInterval)
		defer t.Stop()
		keepaliveTickerChan = t.C
	}

	var reason string
loop:
	for {
		select {
		case <-keepaliveTickerChan:
			written, err := w.Write([]byte(keepAlivePayload))
			if err != nil {
				reason = fmt.Sprintf("write error (keepalive) (%s)", err)
				break loop
			}
			if written != len(keepAlivePayload) {
				reason = "write error (keepalive) (written != len)"
				break loop
			}
			// keepalive written successfully
			flusher.Flush()

		case <-timeoutChan:
			if !writeHeaderEarly { // write header if it wasn't already
				w.WriteHeader(returnStatus)
			}
			if payload != "" {
				// write payload if needed
				written, err := w.Write([]byte(payload))
				if err != nil {
					reason = fmt.Sprintf("write error (response) (%s)", err)
					break loop
				}
				if written != len(payload) {
					reason = "write error (response) (written != len)"
					break loop
				}
			}

			reason = "server timeout"
			break loop
		case <-r.Context().Done():
			reason = fmt.Sprintf("client closed connection (%s)", r.Context().Err())
			// connection is closed, nothing to write...
			break loop
		}
	}

	log.Printf("[%d] Connection closed! Reason: %s. Was open for %s", reqId, reason, time.Since(connectionStarted))
}

func main() {
	// Flags
	bindAddr := flag.String("bind", ":8080", "Specifies the server's address to bind to.")
	flag.DurationVar(&serverTimeout, "timeout", 0, "Defines the duration for the server to respond and subsequently close the connection. Omit to keep the connection open indefinitely.")
	flag.BoolVar(&writeHeaderEarly, "write-header-early", false, "Enables writing headers as soon as the client connects (while keeping the connection open).")
	flag.IntVar(&returnStatus, "status", 200, "Sets the HTTP status code to include in the response.")
	flag.StringVar(&payload, "payload", "", "Specifies the payload to be sent in the response (generates errors when status is set to 204).")
	flag.StringVar(&keepAlivePayload, "keepalive-payload", " ", "Defines the payload for keep-alive data.")
	flag.DurationVar(&keepAliveInterval, "keepalive-interval", 0, "Sets the frequency to write data for keeping the connection alive. Leave unset to disable keep-alive. Requires enabling write-header-early.")
	flag.Parse()

	if keepAliveInterval != 0 && !writeHeaderEarly {
		log.Fatal("keepalive-interval requires write-header-early to be enabled!")
	}
	if returnStatus == 204 && payload != "" {
		log.Fatal("setting both payload and 204 as status does not make sense")
	}
	if returnStatus == 204 && keepAliveInterval != 0 {
		log.Fatal("setting both payload and keepalive-interval does not make sense")
	}

	// Do the magic
	log.Printf("Starting HTTP server on %s\n", *bindAddr)
	err := http.ListenAndServe(*bindAddr, Handler{})
	if err != nil {
		log.Fatal(err)
	}
}
