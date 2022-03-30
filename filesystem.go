package filesystem

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"sort"
)

type FileSystem interface {

	// Open opens the named file for reading.
	// if file not exist, the error is os.ErrNotExist.
	Open(name string) (File, error)

	// Create creates the named file with mode 0666 (before umask), truncating
	// it if it already exists. If successful, methods on the returned
	// File can be used for writing; the associated file descriptor has
	// mode O_RDWR.
	// If there is an error, it will be of type *PathError.
	Create(name string) (File, error)

	// Mkdir creates a new directory with the specified name and permission
	// bits (before umask).
	// If there is an error, it will be of type *PathError.
	Mkdir(name string, perm os.FileMode) error

	// MkdirAll creates a directory named path,
	// along with any necessary parents, and returns nil,
	// or else returns an error.
	// The permission bits perm (before umask) are used for all
	// directories that MkdirAll creates.
	// If path is already a directory, MkdirAll does nothing
	// and returns nil.
	MkdirAll(path string, perm os.FileMode) error

	// Remove removes the named file or (empty) directory.
	// If there is an error, it will be of type *PathError.
	Remove(name string) error

	// RemoveAll removes path and any children it contains.
	// It removes everything it can but returns the first error
	// it encounters. If the path does not exist, RemoveAll
	// returns nil (no error).
	RemoveAll(path string) error

	// Rename renames (moves) oldpath to newpath.
	// If newpath already exists and is not a directory, Rename replaces it.
	// OS-specific restrictions may apply when oldpath and newpath are in different directories.
	// If there is an error, it will be of type *LinkError.
	Rename(oldpath, newpath string) error

	// Sub returns a Filesystem that is backed by the receiver
	Sub(dir string) (FileSystem, error)

	// Stat returns a FileInfo describing the named file.
	Stat(name string) (os.FileInfo, error)

	// Exists checks if the file exists
	Exists(name string) bool

	// IsFile returns true if the given path is a file
	IsFile(name string) bool

	// IsDir returns true if the given path is a directory
	IsDir(name string) bool
}

type File interface {
	fs.File
	io.Writer
}

type ReadDirFS interface {
	FileSystem
	// ReadDir see fs.ReadDirFS
	ReadDir(name string) ([]fs.DirEntry, error)
}

type ReadDirFile interface {
	File
	// ReadDir see fs.ReadDirFile
	ReadDir(n int) ([]fs.DirEntry, error)
}

// ReadDir see fs.ReadDir
func ReadDir(vfs FileSystem, name string) ([]fs.DirEntry, error) {
	if vfs, ok := vfs.(ReadDirFS); ok {
		return vfs.ReadDir(name)
	}

	file, err := vfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dir, ok := file.(ReadDirFile)
	if !ok {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: errors.New("not implemented")}
	}

	list, err := dir.ReadDir(-1)
	sort.Slice(list, func(i, j int) bool { return list[i].Name() < list[j].Name() })
	return list, err
}

type ReadFileFS interface {
	FileSystem
	// ReadFile see fs.ReadFile
	ReadFile(name string) ([]byte, error)
}

// ReadFile see fs.ReadFile
func ReadFile(vfs FileSystem, name string) ([]byte, error) {
	if vfs, ok := vfs.(ReadFileFS); ok {
		return vfs.ReadFile(name)
	}

	file, err := vfs.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var size int
	if info, err := file.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}

	data := make([]byte, 0, size+1)
	for {
		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
		n, err := file.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return data, err
		}
	}
}

type WriteFileFS interface {
	FileSystem
	// WriteFile see os.WriteFile
	WriteFile(name string, data []byte) error
}

// WriteFile see os.WriteFile
func WriteFile(vfs FileSystem, name string, data []byte) error {
	if vfs, ok := vfs.(WriteFileFS); ok {
		return vfs.WriteFile(name, data)
	}

	file, err := vfs.Create(name)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

type OpenFileFs interface {
	FileSystem
	// OpenFile see os.OpenFile
	OpenFile(name string) (File, error)
}

// OpenFile see os.OpenFile
func OpenFile(vfs FileSystem, name string) (File, error) {
	if vfs, ok := vfs.(OpenFileFs); ok {
		return vfs.OpenFile(name)
	}

	return vfs.Open(name)
}
