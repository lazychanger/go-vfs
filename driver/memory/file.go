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

type memFile struct {
	buf *bytes.Buffer

	fi *memFileInfo

	sync.Mutex
}

func (m *memFile) Write(p []byte) (n int, err error) {
	m.Lock()
	defer m.Unlock()
	return m.buf.Write(p)
}

func (m *memFile) Stat() (fs.FileInfo, error) {
	return m.fi, nil
}

func (m *memFile) Read(bytes []byte) (int, error) {
	n, err := m.buf.Read(bytes)

	if n > 0 {
		m.Lock()
		m.fi.size -= int64(n)
		m.Unlock()
	}

	if err != nil {
		return n, err
	}

	return m.buf.Read(bytes)
}

func (m *memFile) Close() error {
	return nil
}

func (m *memFile) Copy() *memFile {
	m.Lock()
	defer m.Unlock()

	data := make([]byte, m.fi.size)

	copy(data, m.buf.Bytes())

	return &memFile{
		buf: bytes.NewBuffer(data),
		fi: &memFileInfo{
			name:  m.fi.name,
			size:  m.fi.size,
			ctime: m.fi.ctime,
		},
	}
}

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
