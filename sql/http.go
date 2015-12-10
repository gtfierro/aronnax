package main

import (
	query "./lang"
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type httpServer struct {
	Port    int
	Backend *mysqlBackend
}

func StartHTTPServer(backend *mysqlBackend, port int) {
	h := &httpServer{Port: port, Backend: backend}
	http.HandleFunc("/query", h.HandleQuery)
	log.Printf("Starting HTTP server on port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

//TODO: currently does not support semicolons embedded inside queries
func (h *httpServer) HandleQuery(w http.ResponseWriter, r *http.Request) {
	var (
		parsed   *query.Query
		parseErr error
		evalErr  error
		docs     []*Document
		encoder  *json.Encoder
	)
	reqBody := bufio.NewReader(r.Body)
	s, err := reqBody.ReadString(';')
	if err != nil {
		w.WriteHeader(400) // Bad Request
		w.Write([]byte("Could not find ';' to terminate query"))
		goto deliver
	}

	// parse query
	parsed, parseErr = h.Backend.Parse(s)
	if parseErr != nil {
		w.WriteHeader(400) // Bad Request
		w.Write([]byte(parseErr.Error()))
		goto deliver
	}

	// eval query
	docs, evalErr = h.Backend.Eval(parsed)
	if evalErr != nil {
		w.WriteHeader(500) // server error
		w.Write([]byte(evalErr.Error()))
		goto deliver
	}

	encoder = json.NewEncoder(w)
	encoder.Encode(docs)
deliver:
	reqBody.Discard(reqBody.Buffered())
	r.Body.Close()
}
