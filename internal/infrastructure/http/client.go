package http

import (
	"io"
	"net/http"
)

// Client is a common interface of http client using in app.
type Client interface {
	Get(url string) (resp *http.Response, err error)
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}
