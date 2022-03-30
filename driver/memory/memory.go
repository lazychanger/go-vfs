package memory

import (
	"errors"
	"github.com/lazychanger/filesystem"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"
	"time"
)

type memFs struct {
	root string

	config *Config

	size uint64

	top *memFs

	files map[string]*memFile
	dirs  map[string]*memFs

	fi *memFileInfo

	sync.RWMutex
}

func (m *memFs) ReadDir(name string) ([]fs.DirEntry, error) {
	dirs := make([]fs.DirEntry, 0)

	node, exist := m.node(name, false)
	if !exist || len(node.dirs) == 0 {
		return dirs, io.EOF
	}

	node.RLock()
	defer node.RUnlock()

	for _, file := range node.files {
		dirs = append(dirs, &memDirEntry{
			fi: file.fi,
		})
	}

	for _, dir := range node.dirs {
		dirs = append(dirs, &memDirEntry{
			fi: dir.fi,
		})
	}

	return dirs, nil
}

func New(config *Config, root string) filesystem.FileSystem {
	if config == nil {
		config = &Config{}
	}

	_, name := dirname(root)

	return &memFs{
		root:   strings.TrimRight(root, "/") + "/",
		config: config,
		files:  make(map[string]*memFile),
		dirs:   make(map[string]*memFs),
		fi: &memFileInfo{
			name:  name,
			size:  0,
			isDir: true,
			ctime: time.Now(),
		},
	}
}

func (m *memFs) Open(name string) (filesystem.File, error) {
	dir, fname := dirname(name)

	node, exist := m.node(dir, false)
	if !exist {
		return nil, &fs.PathError{Op: "open", Path: pathjoin(m.root, name), Err: fs.ErrNotExist}
	}

	return node.open(fname)
}

func (m *memFs) open(name string) (filesystem.File, error) {
	m.RLock()
	defer m.RUnlock()
	if f, ok := m.files[name]; ok {
		return f, nil
	}
	return nil, &fs.PathError{Op: "open", Path: pathjoin(m.root, name), Err: fs.ErrNotExist}
}

func (m *memFs) Create(name string) (filesystem.File, error) {
	dir, fname := dirname(name)

	node, exist := m.node(dir, false)
	if !exist {
		return nil, &fs.PathError{Op: "open", Path: pathjoin(m.root, name), Err: fs.ErrNotExist}
	}

	if f := node.getFile(fname); f != nil {
		return f, nil
	}

	return node.create(fname, nil)
}

func (m *memFs) create(name string, f *memFile) (filesystem.File, error) {
	m.Lock()
	defer m.Unlock()

	if f != nil {
		f.fi.name = name
		f.fi.ctime = time.Now()
		m.files[name] = f
	} else {
		m.files[name] = newMemFile(name, nil, false)
	}

	return m.files[name], nil
}

func (m *memFs) Mkdir(name string, perm fs.FileMode) error {
	dir, dname := dirname(name)

	node, exist := m.node(dir, false)
	if !exist {
		return &fs.PathError{Op: "mkdir", Path: pathjoin(m.root, name), Err: fs.ErrNotExist}
	}

	return node.mkdir(dname)
}

func (m *memFs) mkdir(name string) (err error) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.dirs[name]; ok {
		return &fs.PathError{Op: "mkdir", Path: pathjoin(m.root, name), Err: fs.ErrExist}
	}

	m.dirs[name] = New(m.config, m.root+name+"/").(*memFs)

	return nil
}

func (m *memFs) MkdirAll(path string, perm fs.FileMode) error {
	m.node(path, true)
	return nil
}

func (m *memFs) Remove(name string) error {
	dir, fname := dirname(name)

	node, exist := m.node(dir, false)

	if !exist {
		return &fs.PathError{Op: "removeFile", Path: pathjoin(m.root, name), Err: fs.ErrNotExist}
	}

	if strings.HasSuffix(name, "/") {
		if _, err := node.removeDir(fname); err != nil {
			return err
		}
		return nil
	}

	if _, err := node.removeFile(fname); err == nil {
		return nil
	}

	if _, err := node.removeDir(fname); err != nil {
		return err
	}

	return nil
}

func (m *memFs) removeFile(name string) (*memFile, error) {
	m.Lock()
	defer m.Unlock()

	if f, ok := m.files[name]; ok {
		delete(m.files, name)
		return f, nil
	}
	return nil, &fs.PathError{Op: "removeFile", Path: pathjoin(m.root, name), Err: fs.ErrNotExist}
}

func (m *memFs) RemoveAll(path string) error {
	if path == "/" {
		m.Lock()
		defer m.Unlock()

		m.files = make(map[string]*memFile)
		m.dirs = make(map[string]*memFs)
		m.size = 0
		return nil
	}

	dir, fname := dirname(path)

	node, exist := m.node(dir, false)
	if !exist {
		return nil
	}

	if strings.HasSuffix(path, "/") {
		_, _ = node.removeDir(fname)
		return nil
	}

	if _, err := node.removeFile(fname); err == nil {
		return nil
	}

	_, _ = node.removeDir(fname)

	return nil
}

func (m *memFs) removeDir(name string) (*memFs, error) {

	m.Lock()
	defer m.Unlock()

	if dir, ok := m.dirs[name]; ok {
		delete(m.dirs, name)
		return dir, nil
	}

	return nil, &fs.PathError{Op: "removeDir", Path: pathjoin(m.root, name), Err: fs.ErrNotExist}
}

func (m *memFs) Rename(oldpath, newpath string) error {
	if strings.HasSuffix(oldpath, "/") {
		if err := m.renameDir(oldpath, newpath); err != nil {
			return err
		}
		return nil
	}

	if err := m.renameFile(oldpath, newpath); err == nil {
		return nil
	}

	if err := m.renameDir(oldpath, newpath); err != nil {
		return err
	}

	return nil
}

func (m *memFs) Sub(dir string) (filesystem.FileSystem, error) {
	if dir == "." || dir == ".." {
		return nil, errors.New("invalid sub directory")
	}

	node, exist := m.node(dir, false)

	if !exist {
		return nil, &fs.PathError{Op: "sub", Path: pathjoin(m.root, dir), Err: fs.ErrNotExist}
	}

	return node, nil
}

func (m *memFs) Stat(name string) (fs.FileInfo, error) {
	dir, fname := dirname(name)

	node, exist := m.node(dir, false)
	if !exist {
		return nil, &fs.PathError{Op: "Stat", Path: pathjoin(m.root, name), Err: fs.ErrNotExist}
	}

	if node.existFile(fname) {
		return node.files[fname].fi, nil
	}

	if node.existDir(fname) {
		return node.dirs[fname].fi, nil
	}

	return nil, &fs.PathError{Op: "Stat", Path: pathjoin(m.root, name), Err: fs.ErrNotExist}
}

func (m *memFs) Exists(name string) bool {
	dir, fname := dirname(name)

	node, exist := m.node(dir, false)

	if !exist {
		return false
	}

	if node.existFile(fname) || node.existDir(fname) {
		return true
	}

	return false
}

func (m *memFs) IsFile(name string) bool {
	dir, fname := dirname(name)

	node, exist := m.node(dir, false)
	if !exist {
		return false
	}

	return node.existFile(fname)
}

func (m *memFs) IsDir(name string) bool {
	dir, fname := dirname(name)

	node, exist := m.node(dir, false)
	if !exist {
		return false
	}

	return node.existDir(fname)
}

func (m *memFs) renameFile(oldpath, newpath string) error {
	olddir, oldname := dirname(oldpath)
	onode, oexist := m.node(olddir, false)

	if !oexist {
		return &fs.PathError{Op: "renameFile", Path: pathjoin(m.root, oldpath), Err: fs.ErrNotExist}
	}

	newdir, newname := dirname(newpath)

	nnode, nexist := m.node(newdir, false)
	if !nexist {
		return &fs.PathError{Op: "renameFile", Path: pathjoin(m.root, newdir), Err: fs.ErrNotExist}
	}

	f, err := onode.removeFile(oldname)
	if err != nil {
		return err
	}

	_, _ = nnode.create(newname, f)
	return nil
}

func (m *memFs) renameDir(oldpath, newpath string) error {
	olddir, oldname := dirname(oldpath)

	onode, oexist := m.node(olddir, false)
	if !oexist {
		return &fs.PathError{Op: "renameDir", Path: pathjoin(m.root, oldname), Err: fs.ErrNotExist}
	}

	newdir, newname := dirname(newpath)

	nnode, nexist := m.node(newdir, false)

	if !nexist {
		return &fs.PathError{Op: "renameDir", Path: pathjoin(m.root, newdir), Err: fs.ErrNotExist}
	}

	if nnode.existDir(newname) {
		return &fs.PathError{Op: "renameDir", Path: pathjoin(m.root, newname), Err: fs.ErrExist}
	}

	dir, err := onode.removeDir(oldname)
	if err != nil {
		return err
	}

	dir.Lock()
	dir.root = path.Join(dir.root, "../", newname)
	dir.fi.name = newname
	dir.Unlock()

	nnode.Lock()
	nnode.dirs[newname] = dir
	nnode.Unlock()

	return nil
}

func (m *memFs) existFile(file string) bool {
	m.RLock()
	defer m.RUnlock()
	if _, ok := m.files[file]; ok {
		return true
	}
	return false
}

func (m *memFs) getFile(file string) filesystem.File {
	m.RLock()
	defer m.RUnlock()

	if f, ok := m.files[file]; ok {
		return f

	} else {
		return nil
	}
}

func (m *memFs) existDir(dir string) bool {
	m.RLock()
	defer m.RUnlock()
	if _, ok := m.dirs[dir]; ok {
		return true
	}
	return false
}

func (m *memFs) node(dir string, autoCreate bool) (fs *memFs, exists bool) {
	dir = strings.TrimRight(dir, "/")
	if dir == "" {
		return m, true
	}

	dirs := strings.Split(dir, "/")

	node := m
	top := m.top

	if top == nil {
		top = m
	}

	if strings.HasPrefix(dir, "/") {
		node = top
	}

	for i := 0; i < len(dirs); i++ {
		if dirs[i] == "" {
			continue
		}
		node.RLock()
		if child, ok := node.dirs[dirs[i]]; ok {
			node.RUnlock()
			node = child
			continue
		}

		node.RUnlock()
		if !autoCreate {
			return nil, false
		}

		child := New(m.config, node.root+dirs[i]).(*memFs)
		child.top = top
		node.Lock()

		node.dirs[dirs[i]] = child

		node.Unlock()

		node = child
	}

	return node, true
}

func dirname(path string) (dir string, name string) {
	dirs := strings.Split(strings.TrimRight(path, "/"), "/")

	return strings.Join(dirs[:len(dirs)-1], "/"), dirs[len(dirs)-1]
}

func pathjoin(root, dir string) string {
	if strings.HasPrefix(dir, "/") {
		return dir
	}
	return path.Join(root, dir)
}
