package mock

import "errors"

// FailReader is mock that imitates errors.
type FailReader struct{}

// Read always returns error.
func (f *FailReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}
