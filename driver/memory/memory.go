package memory

import (
	"github.com/lazychanger/filesystem"
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

	fi fs.FileInfo

	sync.RWMutex
}

func New(config *Config, root string) filesystem.FileSystem {
	return &memFs{
		root:   strings.TrimRight(root, "/") + "/",
		config: config,
		files:  make(map[string]*memFile),
		dirs:   make(map[string]*memFs),
		fi: &memFileInfo{
			name:  "",
			size:  0,
			isDir: true,
			ctime: time.Now(),
		},
	}
}

func (m *memFs) Open(name string) (filesystem.File, error) {
	dir, fname := m.path(name)

	node, exist := m.node(dir, false)
	if !exist {
		return nil, &fs.PathError{Op: "open", Path: path.Join(m.root, fname), Err: fs.ErrNotExist}
	}
	return node.open(fname)
}

func (m *memFs) open(name string) (filesystem.File, error) {
	m.RLock()
	defer m.RUnlock()

	if f, ok := m.files[name]; ok {
		return f, nil
	}

	return nil, &fs.PathError{Op: "open", Path: path.Join(m.root, name), Err: fs.ErrNotExist}
}

func (m *memFs) Create(name string) (filesystem.File, error) {
	dir, fname := m.path(name)

	node, exist := m.node(dir, false)
	if !exist {
		return nil, &fs.PathError{Op: "open", Path: m.root, Err: fs.ErrNotExist}
	}

	if !m.existFile(name) {
		return nil, &fs.PathError{Op: "open", Path: path.Join(m.root, fname), Err: fs.ErrExist}
	}

	return node.create(fname, nil)
}

func (m *memFs) create(name string, f *memFile) (filesystem.File, error) {
	m.Lock()
	defer m.Unlock()

	if f != nil {
		m.files[name] = f
	} else {
		m.files[name] = newMemFile(name, nil, false)
	}

	return m.files[name], nil
}

func (m *memFs) Mkdir(name string, perm fs.FileMode) error {
	dir, dirname := m.path(name)

	node, exist := m.node(dir, false)
	if !exist {
		return &fs.PathError{Op: "mkdir", Path: path.Join(m.root, dirname), Err: fs.ErrNotExist}
	}

	return node.mkdir(dirname)
}

func (m *memFs) mkdir(name string) (err error) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.dirs[name]; ok {
		return &fs.PathError{Op: "mkdir", Path: path.Join(m.root, name), Err: fs.ErrExist}
	}

	m.dirs[name] = New(m.config, m.root+name+"/").(*memFs)

	return nil
}

func (m *memFs) MkdirAll(path string, perm fs.FileMode) error {
	m.node(path, true)
	return nil
}

func (m *memFs) Remove(name string) error {
	dir, fname := m.path(name)

	node, exist := m.node(dir, false)
	if !exist {
		return &fs.PathError{Op: "removeFile", Path: path.Join(m.root, fname), Err: fs.ErrNotExist}
	}

	if _, err := node.removeFile(fname); err != nil {
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

	return nil, &fs.PathError{Op: "removeFile", Path: path.Join(m.root, name), Err: fs.ErrNotExist}
}

func (m *memFs) RemoveAll(path string) error {
	dir, fname := m.path(path)

	node, exist := m.node(dir, false)
	if !exist {
		return nil
	}

	_, _ = node.removeFile(fname)

	if _, err := node.removeDir(fname); err != nil {
		return err
	}

	return nil
}

func (m *memFs) removeDir(name string) (*memFs, error) {

	m.Lock()
	defer m.Unlock()

	if dir, ok := m.dirs[name]; ok {
		delete(m.dirs, name)
		return dir, nil
	}

	return nil, &fs.PathError{Op: "removeDir", Path: path.Join(m.root, name), Err: fs.ErrNotExist}
}

func (m *memFs) Rename(oldpath, newpath string) error {
	renameFileErr := m.renameFile(oldpath, newpath)

	renameDirErr := m.renameDir(oldpath, newpath)

	if renameFileErr != nil && renameDirErr != nil {
		return renameDirErr
	}
	return nil
}

func (m *memFs) Sub(name string) (filesystem.FileSystem, error) {

	node, exist := m.node(name, false)
	if !exist {
		return nil, &fs.PathError{Op: "sub", Path: path.Join(node.root), Err: fs.ErrNotExist}
	}

	return node, nil
}

func (m *memFs) Stat(name string) (fs.FileInfo, error) {
	dir, fname := m.path(name)

	node, exist := m.node(dir, false)
	if !exist {
		return nil, &fs.PathError{Op: "Stat", Path: path.Join(node.root), Err: fs.ErrNotExist}
	}

	if node.existFile(fname) {
		return node.files[fname].fi, nil
	}

	if node.existDir(fname) {
		return node.dirs[fname].fi, nil
	}

	return nil, &fs.PathError{Op: "Stat", Path: path.Join(node.root, fname), Err: fs.ErrNotExist}
}

func (m *memFs) renameFile(oldpath, newpath string) error {
	olddir, oldname := m.path(oldpath)

	onode, exist := m.node(olddir, false)
	if !exist {
		return &fs.PathError{Op: "renameFile", Path: path.Join(oldpath), Err: fs.ErrNotExist}
	}

	newdir, newname := m.path(newpath)

	nnode, _ := m.node(newdir, true)

	if !nnode.existFile(newname) {
		return &fs.PathError{Op: "renameFile", Path: path.Join(nnode.root, newname), Err: fs.ErrExist}
	}

	f, err := onode.removeFile(oldname)
	if err != nil {
		return err
	}

	if _, err := nnode.create(newname, f); err != nil {
		return err
	}
	return nil
}

func (m *memFs) renameDir(oldpath, newpath string) error {
	olddir, oldname := m.path(oldpath)

	onode, exist := m.node(olddir, false)
	if !exist {
		return &fs.PathError{Op: "rename", Path: path.Join(onode.root, oldname), Err: fs.ErrNotExist}
	}

	newdir, newname := m.path(newpath)

	nnode, _ := m.node(newdir, true)

	if nnode.existDir(newname) {
		return &fs.PathError{Op: "rename", Path: path.Join(nnode.root, newname), Err: fs.ErrExist}
	}

	dir, err := onode.removeDir(oldname)
	if err != nil {
		return err
	}

	dir.Lock()
	dir.root = path.Join(dir.root, "../", newname)
	dir.Unlock()

	nnode.RLock()
	nnode.dirs[newname] = dir
	nnode.Unlock()

	return nil
}

func (m *memFs) existFile(file string) bool {
	m.RLock()
	defer m.Unlock()
	if _, ok := m.files[file]; ok {
		return true
	}
	return false
}

func (m *memFs) existDir(dir string) bool {
	m.RLock()
	defer m.Unlock()
	if _, ok := m.dirs[dir]; ok {
		return true
	}
	return false
}

func (m *memFs) path(path string) (dir string, name string) {
	dirs := strings.Split(strings.TrimRight(path, "/"), "/")

	return strings.Join(dirs[:len(dirs)-1], "/"), dirs[len(dirs)-1]
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
		node.RLock()
		if child, ok := node.dirs[dirs[i]]; ok {
			node = child
			node.RUnlock()
			continue
		}

		node.RUnlock()
		if !autoCreate {
			return nil, false
		}

		child := New(m.config, m.root+dirs[i]).(*memFs)
		child.top = top
		node.Lock()

		node.dirs[dirs[i]] = child

		node.Unlock()

		node = child
	}

	return node, true
}
