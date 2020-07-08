package os

import (
	"io"
	"os"
)

// A FileMode represents a file's mode and permission bits.
// The bits have the same definition on all systems, so that
// information about files can be moved from one system
// to another portably. Not all bits apply to all systems.
// The only required bit is ModeDir for directories.
type FileMode uint32

// The defined file mode bits are the most significant bits of the FileMode.
// The nine least-significant bits are the standard Unix rwxrwxrwx permissions.
// The values of these bits should be considered part of the public API and
// may be used in wire protocols or disk representations: they must not be
// changed, although new bits might be added.
const (
	// The single letters are the abbreviations
	// used by the String method's formatting.
	ModeDir        FileMode = 1 << (32 - 1 - iota) // d: is a directory
	ModeAppend                                     // a: append-only
	ModeExclusive                                  // l: exclusive use
	ModeTemporary                                  // T: temporary file; Plan 9 only
	ModeSymlink                                    // L: symbolic link
	ModeDevice                                     // D: device file
	ModeNamedPipe                                  // p: named pipe (FIFO)
	ModeSocket                                     // S: Unix domain socket
	ModeSetuid                                     // u: setuid
	ModeSetgid                                     // g: setgid
	ModeCharDevice                                 // c: Unix character device, when ModeDevice is set
	ModeSticky                                     // t: sticky
	ModeIrregular                                  // ?: non-regular file; nothing else is known about this file

	// Mask for the type bits. For regular files, none will be set.
	ModeType = ModeDir | ModeSymlink | ModeNamedPipe | ModeSocket | ModeDevice | ModeCharDevice | ModeIrregular

	ModePerm FileMode = 0777 // Unix permission bits
)

// FileSystem is the interface that wraps common methods for working with filesystem.
type FileSystem interface {
	// Open opens the named file for reading.
	Open(name string) (File, error)
	// Stat returns a FileInfo describing the named file.
	Stat(name string) (os.FileInfo, error)
	// IsNotExist returns a boolean indicating whether the error is known to
	// report that a file or directory does not exist.
	IsNotExist(err error) bool
	// MkdirAll creates a directory named path.
	MkdirAll(path string, perm FileMode) error
}

// File is the interface that wraps common methods for working with files.
type File interface {
	io.Closer
	io.Reader
	io.ReaderAt
	io.Seeker
	// Stat returns a FileInfo describing the named file.
	Stat() (os.FileInfo, error)
}

// FS is the implementation of FileSystem interface as a real OS.
type FS struct{}

// Open opens the named file for reading. If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func (FS) Open(name string) (File, error) {
	return os.Open(name)
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func (FS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// IsNotExist returns a boolean indicating whether the error is known to
// report that a file or directory does not exist. It is satisfied by
// ErrNotExist as well as some syscall errors.
func (FS) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// MkdirAll creates a directory named path,
// along with any necessary parents, and returns nil,
// or else returns an error.
// The permission bits perm (before umask) are used for all
// directories that MkdirAll creates.
// If path is already a directory, MkdirAll does nothing
// and returns nil.
func (FS) MkdirAll(path string, perm FileMode) error {
	return os.MkdirAll(path, os.FileMode(perm))
}
