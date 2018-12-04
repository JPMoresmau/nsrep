package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"

	item "github.com/JPMoresmau/metarep/item"
)

func writeError(w http.ResponseWriter, err error) {
	log.Println(err)
	log.Println(err == nil)
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
	store     item.Store
	secondary item.Store
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
		if err == nil && sh.secondary != nil {
			err = sh.secondary.Write(it)
		}
	case "DELETE":
		err = sh.store.Delete(id)
		if err == nil {
			if sh.secondary != nil {
				err = sh.secondary.Delete(id)
			}
			if err == nil {
				writeStatus(w, "", http.StatusNoContent)
				return
			}
		}
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
	if req.Method == "GET" && it.IsEmpty() {
		writeStatus(w, resp, http.StatusNotFound)
		return
	}
	writeOK(w, resp)

}

// HistoryHandler is the handler with an history item store
type HistoryHandler struct {
	store item.HistoryStore
}

func (sh *HistoryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var resp string
	path := req.URL.Path
	id := strings.SplitAfter(path, "history/")[1]
	if id == "" {
		writeStatus(w, `{"error":"no id"}`, http.StatusBadRequest)
		return
	}
	ls := req.URL.Query()["limit"]
	var limit = 100
	if len(ls) > 0 {
		l, err := strconv.Atoi(ls[0])
		if err == nil && l > 0 {
			limit = l
		}
	}
	var its = []item.ItemStatus{}
	var err error
	switch req.Method {
	case "GET":
		its, err = sh.store.History(id, limit)
	}
	if err != nil {
		writeError(w, err)
		return
	}
	b, err := json.Marshal(its)
	if err != nil {
		writeError(w, err)
		return
	}
	resp = fmt.Sprintf("%s", b)
	if req.Method == "GET" && len(its) == 0 {
		writeStatus(w, resp, http.StatusNotFound)
		return
	}
	writeOK(w, resp)
}

func startServer(port int, store item.Store, secondary item.Store) *http.Server {
	mux := http.NewServeMux()
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	mux.Handle("/items/", &StoreHandler{store, secondary})
	if h, ok := store.(item.HistoryStore); ok {
		mux.Handle("/history/", &HistoryHandler{h})
	}

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
	srv := startServer(8080, store, nil)
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
