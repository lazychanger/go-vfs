package os

import (
	"errors"
	"github.com/lazychanger/filesystem"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
)

// fileSystem is the file system implementation for the os package.
type fileSystem struct {
	config *Config

	top *fileSystem

	parent *fileSystem

	children map[string]*fileSystem
}

func New(config *Config) (filesystem.FileSystem, error) {
	if config.Root == "" {
		return nil, fs.ErrInvalid
	}

	if config.Root == "/" {
		return nil, fs.ErrNotExist
	}

	if !strings.HasPrefix(config.Root, "/") {
		return nil, errors.New("root must be absolute path")
	}

	vfs := &fileSystem{
		config: config,
	}

	if err := vfs.init(); err != nil {
		return nil, err
	}

	return &fileSystem{
		config: config,
	}, nil
}

func (vfs *fileSystem) init() error {
	vfs.config.Root = strings.TrimRight(vfs.config.Root, "/")

	dirs := strings.Split(vfs.config.Root, "/")

	dir, _ := strings.Join(dirs[:len(dirs)-1], "/"), dirs[len(dirs)-1]

	if _, err := os.Stat(dir); err != nil {
		return err
	}
	log.Println(vfs.config.Root)
	return os.MkdirAll(vfs.config.Root, 0755)
}

func (vfs *fileSystem) Open(name string) (filesystem.File, error) {
	return os.Open(vfs.path(name))
}

func (vfs *fileSystem) Create(name string) (filesystem.File, error) {
	return os.Create(vfs.path(name))
}

func (vfs *fileSystem) Mkdir(name string, perm fs.FileMode) error {
	return os.Mkdir(vfs.path(name), perm)
}

func (vfs *fileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(vfs.path(path), perm)
}

func (vfs *fileSystem) Remove(name string) error {
	return os.Remove(vfs.path(name))
}

func (vfs *fileSystem) RemoveAll(path string) error {
	return os.RemoveAll(vfs.path(path))
}

func (vfs *fileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(vfs.path(oldpath), vfs.path(newpath))
}

func (vfs *fileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(path.Join(vfs.config.Root, name))
}

func (vfs *fileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(vfs.path(name))
}

func (vfs *fileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(vfs.path(name))
}

func (vfs *fileSystem) WriteFile(name string, data []byte) error {
	return os.WriteFile(vfs.path(name), data, 0755)
}

func (vfs *fileSystem) OpenFile(name string) (filesystem.File, error) {
	return os.OpenFile(vfs.path(name), os.O_RDONLY, 0755)
}

func (vfs *fileSystem) Sub(dir string) (filesystem.FileSystem, error) {
	if dir == "." || dir == ".." {
		return nil, errors.New("invalid sub directory")
	}

	if !vfs.IsDir(dir) {
		return nil, &fs.PathError{Op: "sub", Path: dir, Err: fs.ErrNotExist}
	}

	return &fileSystem{
		config: &Config{Root: vfs.path(dir)},
	}, nil
}

func (vfs *fileSystem) Exists(name string) bool {
	_, err := os.Stat(vfs.path(name))
	return err == nil
}

func (vfs *fileSystem) IsFile(name string) bool {
	info, err := os.Stat(vfs.path(name))
	return err == nil && !info.IsDir()
}

func (vfs *fileSystem) IsDir(name string) bool {
	info, err := os.Stat(vfs.path(name))
	return err == nil && info.IsDir()
}

func (vfs *fileSystem) path(p string) string {
	return path.Join(vfs.config.Root, p)
}
