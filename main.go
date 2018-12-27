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
	"time"

	item "github.com/JPMoresmau/nsrep/item"
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
		if h2, ok2 := sh.store.(item.SearchStore); ok2 {
			err = item.DeleteTree(id, []item.Store{sh.store, sh.secondary}, h2)
			if err == nil {
				writeStatus(w, "", http.StatusNoContent)
				return
			}
		} else if h2, ok2 := sh.secondary.(item.SearchStore); ok2 {
			err = item.DeleteTree(id, []item.Store{sh.store, sh.secondary}, h2)
			if err == nil {
				writeStatus(w, "", http.StatusNoContent)
				return
			}
		} else {
			err = sh.store.Delete(id)
			if err == nil {
				if sh.secondary != nil {
					go sh.secondary.Delete(id)
				}
				if err == nil {
					writeStatus(w, "", http.StatusNoContent)
					return
				}
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
	limit := positiveIntParam(req, "limit", 100)
	var its = []item.Status{}
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

// SearchHandler is the handler with an history item store
type SearchHandler struct {
	store item.SearchStore
}

func positiveIntParam(req *http.Request, name string, def int) int {
	ls := req.URL.Query()[name]
	var val = def
	if len(ls) > 0 {
		l, err := strconv.Atoi(ls[0])
		if err == nil && l > 0 {
			val = l
		}
	}
	return val
}

func (sh *SearchHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var resp string
	var queries = req.URL.Query()["query"]
	if len(queries) == 0 {
		writeStatus(w, `{"error":"no query"}`, http.StatusBadRequest)
		return
	}
	var query = queries[0]
	var from = positiveIntParam(req, "from", 0)
	var length = positiveIntParam(req, "length", 10)
	var its = []item.Score{}
	var err error
	switch req.Method {
	case "GET":
		its, err = sh.store.Search(item.Page(item.NewQuery(query), from, length))
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
	writeOK(w, resp)
}

func startServer(port int, store item.Store, secondary item.Store) *http.Server {
	mux := http.NewServeMux()
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	mux.Handle("/items/", &StoreHandler{store, secondary})
	if h, ok := store.(item.HistoryStore); ok {
		mux.Handle("/history/", &HistoryHandler{h})
	} else if secondary != nil {
		if h2, ok2 := secondary.(item.HistoryStore); ok2 {
			mux.Handle("/history/", &HistoryHandler{h2})
		}
	}
	if h, ok := store.(item.SearchStore); ok {
		mux.Handle("/search", &SearchHandler{h})
	} else if secondary != nil {
		if h2, ok2 := secondary.(item.SearchStore); ok2 {
			mux.Handle("/search", &SearchHandler{h2})
		}
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

type storeCreate func() (item.Store, error)

func waitForStore(sc storeCreate) (item.Store, error) {
	return reallyWaitForStore(sc, 6, 1)
}

func reallyWaitForStore(sc storeCreate, nb int, delay time.Duration) (item.Store, error) {
	st, err := sc()
	if err == nil {
		return st, nil
	}
	if nb == 0 {
		return st, err
	}
	time.Sleep(delay * time.Second)
	return reallyWaitForStore(sc, nb-1, delay*2)
}

func main() {
	app := os.Getenv("NSREP_CONFIG_FILE")
	if len(app) == 0 {
		app = "application.yaml"
	}
	log.Printf("Reading configuration from %s\n", app)
	c, err := ReadFileConfig(app)
	if err != nil {
		log.Panicf("Cannot parse application.yaml: %s \n%v", err.Error(), err)
		return
	}
	var cqlCreate storeCreate = func() (item.Store, error) { return item.NewCqlStore(c.Cassandra) }
	store, err := waitForStore(cqlCreate)
	if err != nil {
		log.Panicf("Cannot connect to cassandra: %s \n%v", err.Error(), err)
		return
	}
	log.Println("Connected to Cassandra")
	var esCreate storeCreate = func() (item.Store, error) { return item.NewElasticStore(c.Elastic) }
	secondary, err := waitForStore(esCreate)
	if err != nil {
		log.Panicf("Cannot connect to elastic: %s \n%v", err.Error(), err)
		return
	}
	log.Println("Connected to Elastic")
	srv := startServer(c.Port, store, secondary)
	log.Printf("Server started on port %d\n", c.Port)
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
