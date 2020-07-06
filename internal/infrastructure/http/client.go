package http

import "net/http"

// Client is a common interface of http client using in app.
type Client interface {
	Get(url string) (resp *http.Response, err error)
}
