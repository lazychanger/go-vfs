package memory

import (
	"bytes"
	"io/fs"
	"sync"
	"time"
)

func newMemFile(name string, buf []byte, isDir bool) *memFile {
	return &memFile{
		buf: bytes.NewBuffer(buf),
		fi: &memFileInfo{
			name:  name,
			isDir: isDir,
			ctime: time.Now(),
		},
	}
}

// memFile is a memory file implementation.
// implements io.Writer, io.Reader, io.Closer, fs.FileInfo
type memFile struct {
	buf *bytes.Buffer

	fi *memFileInfo

	sync.Mutex
}

func (m *memFile) Write(p []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	n, err = m.buf.Write(p)
	m.fi.ctime = time.Now()
	m.fi.size += int64(n)
	return
}

func (m *memFile) Stat() (fs.FileInfo, error) {
	return m.fi, nil
}

func (m *memFile) Read(bytes []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	n, err = m.buf.Read(bytes)
	m.fi.size -= int64(n)
	m.fi.ctime = time.Now()

	return n, err
}

func (m *memFile) Close() error {
	return nil
}

//func (m *memFile) Copy() *memFile {
//	m.Lock()
//	defer m.Unlock()
//
//	data := make([]byte, m.fi.size)
//
//	copy(data, m.buf.Bytes())
//
//	return &memFile{
//		buf: bytes.NewBuffer(data),
//		fi: &memFileInfo{
//			name:  m.fi.name,
//			size:  m.fi.size,
//			ctime: m.fi.ctime,
//		},
//	}
//}

// memDirEntry see fs.FileInfo
type memFileInfo struct {
	name  string
	size  int64
	isDir bool
	ctime time.Time
}

func (m *memFileInfo) Name() string {
	return m.name
}

func (m *memFileInfo) Size() int64 {
	return m.size
}

func (m *memFileInfo) Mode() fs.FileMode {
	return fs.FileMode(0)
}

func (m *memFileInfo) ModTime() time.Time {
	return m.ctime
}

func (m *memFileInfo) IsDir() bool {
	return m.isDir
}

func (m *memFileInfo) Sys() any {
	return nil
}

// memDirEntry see fs.DirEntry
type memDirEntry struct {
	fi fs.FileInfo
}

func (m *memDirEntry) Name() string {
	return m.fi.Name()
}

func (m *memDirEntry) IsDir() bool {
	return m.fi.IsDir()
}

func (m *memDirEntry) Type() fs.FileMode {
	return m.fi.Mode()
}

func (m *memDirEntry) Info() (fs.FileInfo, error) {
	return m.fi, nil
}
