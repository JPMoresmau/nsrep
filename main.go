package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	item "github.com/JPMoresmau/metarep/item"
)

func writeError(w http.ResponseWriter, err error) {
	log.Println(err)
	resp := fmt.Sprintf(`{"error":"%s"}`, err.Error())
	writeStatus(w, resp, http.StatusInternalServerError)
}

func writeOK(w http.ResponseWriter, content string) {
	writeStatus(w, content, http.StatusOK)
}

func writeStatus(w http.ResponseWriter, content string, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	io.WriteString(w, content)
}

// StoreHandler is the handler with an item store
type StoreHandler struct {
	store item.Store
}

func (sh *StoreHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var resp string
	path := req.URL.Path
	id := strings.SplitAfter(path, "items/")[1]
	if id == "" {
		writeStatus(w, `{"error":"no id"}`, http.StatusBadRequest)
		return
	}
	it := item.Item{}
	var err error
	switch req.Method {
	case "GET":
		it, err = sh.store.Read(id)
	case "POST":
		err := json.NewDecoder(req.Body).Decode(&it)
		if err != nil {
			writeError(w, err)
			return
		}
		it.ID = id
		err = sh.store.Write(it)
	case "DELETE":
		it, err = sh.store.Delete(id)
	default:
		err = fmt.Errorf("Method %s not supported", req.Method)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	b, err := json.Marshal(it)
	if err != nil {
		writeError(w, err)
		return
	}
	resp = fmt.Sprintf("%s", b)
	if it.IsEmpty() {
		writeStatus(w, resp, http.StatusNotFound)
		return
	}
	writeOK(w, resp)

}

func startServer(port int, store item.Store) *http.Server {
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	http.Handle("/items/", &StoreHandler{store})
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()
	return srv
}

func stopServer(srv *http.Server) {
	if err := srv.Shutdown(nil); err != nil {
		log.Printf("HTTP server Shutdown: %v", err)
	}
}

func main() {
	store := item.NewLocalStore()
	srv := startServer(8080, store)
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		stopServer(srv)
		close(idleConnsClosed)
	}()
	<-idleConnsClosed
}
