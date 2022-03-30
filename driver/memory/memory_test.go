package memory

import (
	"bytes"
	"fmt"
	"github.com/lazychanger/filesystem/tests"
	"github.com/stretchr/testify/assert"
	"io"
	"io/fs"
	"testing"
)

func TestNew(t *testing.T) {
	config := &Config{}

	assert.Equal(t, Driver, config.Driver())
	query := config.Encode()

	assert.Equal(t, query.Encode(), "maxsize=0")

	query.Set("maxsize", "10")

	assert.NoError(t, config.Decode(query))

	assert.Equal(t, int64(10), config.MaxSize)

	vfs := New(config, "/")
	assert.True(t, vfs != nil)

	vfs2 := New(nil, "/")
	assert.True(t, vfs2 != nil)

	f, err := vfs.Create("test.txt")
	assert.NoError(t, err)

	testBytes := bytes.NewBufferString("test")

	n, err := f.Write(testBytes.Bytes())
	assert.NoError(t, err)

	defer f.Close()

	assert.Equal(t, testBytes.Len(), n)

	fi, err := f.Stat()
	assert.NoError(t, err)

	assert.Equal(t, fi.Size(), int64(testBytes.Len()))
	assert.Equal(t, fi.Name(), "test.txt")
	assert.Equal(t, fi.IsDir(), false)
	assert.NotNil(t, fi.ModTime())
	assert.Nil(t, fi.Sys())
	assert.Equal(t, fi.Mode(), fs.FileMode(0))

	fb, _ := io.ReadAll(f)
	assert.Equal(t, string(fb), testBytes.String())

}

func TestMemFs(t *testing.T) {
	tests.TestDriver(t, fmt.Sprintf("memory:///?maxsize=%d", 2>>10))
}
