package tests

import (
	"github.com/lazychanger/filesystem"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDriver(t *testing.T, dsn string) {
	var (
		vfs filesystem.FileSystem
		err error
	)
	vfs, err = filesystem.Open(dsn)
	if err != nil {
		t.Error(err)
		return
	}

	err = vfs.Mkdir("/test1/noexist", 0777)

	assert.Equal(t, err != nil, true, "mkdir use not exist dir path. error is not exists")

	err = vfs.Mkdir("/test1", 0777)

	assert.Equal(t, err == nil, true, "mkdir use root dir path. error is nil")

	err = vfs.MkdirAll("/test2/test/test3", 0777)
	assert.Equal(t, err == nil, true, "mkdir-all will not error")

	_, err = vfs.Create("/test2/test.txt")

	assert.Equal(t, err == nil, true, "mkdir-all will not error")

}
