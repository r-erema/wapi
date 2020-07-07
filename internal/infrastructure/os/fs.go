package os

import (
	"io"
	"os"
)

type FileSystem interface {
	Open(name string) (File, error)
	Stat(name string) (os.FileInfo, error)
	IsNotExist(err error) bool
	MkdirAll(path string, perm os.FileMode) error
}

type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	Stat() (os.FileInfo, error)
}

type FS struct{}

func (FS) Open(name string) (File, error) {
	return os.Open(name)
}
func (FS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
func (FS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}
func (FS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
