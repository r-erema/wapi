package mock

import "errors"

type FailReader struct{}

func (f *FailReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}
