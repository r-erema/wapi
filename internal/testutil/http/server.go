package http

import (
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

// New returns httptest.Server with router
// keys in handlers map represent paths, values represent http.Handler.
func New(handlers map[string]http.Handler) *httptest.Server {
	router := mux.NewRouter().StrictSlash(true)
	for path, handler := range handlers {
		router.Handle(path, handler)
	}
	return httptest.NewServer(router)
}
