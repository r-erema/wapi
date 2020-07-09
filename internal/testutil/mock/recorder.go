package mock

import (
	"errors"
	"net/http"
	"net/http/httptest"
)

// FailReader is mock that imitates writing errors.
type FailResponseRecorder struct {
	recorder httptest.ResponseRecorder
}

// NewFailResponseRecorder creates FailReader object.
func NewFailResponseRecorder(recorder *httptest.ResponseRecorder) *FailResponseRecorder {
	return &FailResponseRecorder{recorder: *recorder}
}

// Header implements http.ResponseWriter. It returns the response
// headers to mutate within a handler.
func (f *FailResponseRecorder) Header() http.Header {
	return f.recorder.Header()
}

// Write never writes anything and always returns error.
func (f *FailResponseRecorder) Write(bytes []byte) (int, error) {
	return 0, errors.New("writing error")
}

// WriteHeader implements http.ResponseWriter.
func (f *FailResponseRecorder) WriteHeader(statusCode int) {
	f.recorder.WriteHeader(statusCode)
}

// Status returns status code.
func (f *FailResponseRecorder) Status() int {
	return f.recorder.Code
}
