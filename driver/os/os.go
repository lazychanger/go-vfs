package os

import (
	"github.com/lazychanger/filesystem"
	"io/fs"
	"os"
	"path"
)

type FileSystem struct {
	config *Config
}

func New(config *Config) (filesystem.FileSystem, error) {
	if config.Root == "" {
		return nil, fs.ErrInvalid
	}

	return &FileSystem{
		config: config,
	}, nil
}

func (vfs *FileSystem) Open(name string) (filesystem.File, error) {
	return os.Open(vfs.path(name))
}

func (vfs *FileSystem) Create(name string) (filesystem.File, error) {
	return os.Create(vfs.path(name))
}

func (vfs *FileSystem) Mkdir(name string, perm fs.FileMode) error {
	return os.Mkdir(vfs.path(name), perm)
}

func (vfs *FileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(vfs.path(path), perm)
}

func (vfs *FileSystem) Remove(name string) error {
	return os.Remove(vfs.path(name))
}

func (vfs *FileSystem) RemoveAll(path string) error {
	return os.RemoveAll(vfs.path(path))
}

func (vfs *FileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(vfs.path(oldpath), vfs.path(newpath))
}

func (vfs *FileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(path.Join(vfs.config.Root, name))
}

func (vfs *FileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(vfs.path(name))
}

func (vfs *FileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(vfs.path(name))
}

func (vfs *FileSystem) Sub(dir string) (filesystem.FileSystem, error) {
	return &FileSystem{
		config: &Config{Root: vfs.path(dir)},
	}, nil
}

func (vfs *FileSystem) path(p string) string {
	return path.Join(vfs.config.Root, p)
}
