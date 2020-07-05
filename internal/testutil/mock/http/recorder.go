package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

type FailResponseRecorder struct {
	recorder httptest.ResponseRecorder
}

func NewFailResponseRecorder(recorder *httptest.ResponseRecorder) *FailResponseRecorder {
	return &FailResponseRecorder{recorder: *recorder}
}

func (f *FailResponseRecorder) Header() http.Header {
	return f.recorder.Header()
}

func (f *FailResponseRecorder) Write(bytes []byte) (int, error) {
	return len(bytes), fmt.Errorf("writing error")
}

func (f *FailResponseRecorder) WriteHeader(statusCode int) {
	f.recorder.WriteHeader(statusCode)
}

func (f *FailResponseRecorder) Status() int {
	return f.recorder.Code
}
