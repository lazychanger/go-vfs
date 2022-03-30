package os

import (
	"fmt"
	"github.com/lazychanger/filesystem/tests"
	"github.com/stretchr/testify/assert"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {

	conf := &Config{
		Root: "",
	}

	assert.Equal(t, conf.Driver(), Driver)
	assert.Equal(t, conf.Root, "")
	assert.Equal(t, conf.Path(), "")

	_, err := New(conf)
	assert.ErrorIs(t, err, fs.ErrInvalid)

	conf.Root = tmpDir()
	_, err = New(conf)
	assert.NoError(t, err)

	conf.Root = "/"
	_, err = New(conf)
	assert.ErrorIs(t, err, fs.ErrNotExist)

	conf.Root = strings.TrimLeft(tmpDir(), "/")
	_, err = New(conf)
	assert.Errorf(t, err, "root must be absolute path")

	noExistPath := path.Join(tmpDir(), "test/test")
	conf.Root = noExistPath
	_, err = New(conf)
	assert.Error(t, err)

}

func TestOsFs(t *testing.T) {
	tests.TestDriver(t, fmt.Sprintf("os://%s/", tmpDir()))
}

func tmpDir() string {

	wd, _ := os.Getwd()
	return path.Join(wd, "../../tmp")
}
