package nona

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var Logs = map[string]func(config json.RawMessage) (Log, error){
	"webhook": NewWebhook,
}

var Stores = map[string]func(config json.RawMessage) (Store, error){
	"json": NewJSON,
}

type Config struct{}

type Log interface {
	Handle(string, *http.Request) error
}

type Store interface {
	Get(key string) (url string, exists bool, err error)
}

type Server struct {
	r *mux.Router
	s Store
	l Log
	c Config
}

// NewServer returns a new Server.
func NewServer(s Store, l Log, c Config) (*Server, error) {
	s2 := &Server{mux.NewRouter(), s, l, c}
	s2.r.HandleFunc("/{key}", s2.get).Methods("GET")
	return s2, nil
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	url, exists, err := s.s.Get(key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("[store] %s: %s", key, err)
		return
	}
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	go func() {
		if err := s.l.Handle(key, r); err != nil {
			log.Printf("[log] %s: %s", key, err)
			return
		}
	}()
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusFound)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.r.ServeHTTP(w, r) }
