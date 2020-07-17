package http

import (
	"net/http/httptest"

	"github.com/r-erema/wapi/internal/http"

	"github.com/gorilla/mux"
)

// New returns httptest.Server with router
// keys in handlers map represent paths, values represent http.Handler.
func New(handlers map[string]http.AppHTTPHandler) *httptest.Server {
	router := mux.NewRouter().StrictSlash(true)
	for path, handler := range handlers {
		router.Handle(path, http.AppHandlerRunner{H: handler})
	}
	return httptest.NewServer(router)
}
