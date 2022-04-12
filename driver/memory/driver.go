package memory

import (
	"github.com/lazychanger/go-vfs"
	"net/url"
)

const Driver = "memory"

func init() {
	filesystem.RegisterDriver(Driver, &fsDriver{})
}

type fsDriver struct {
}

func (m *fsDriver) Open(uri *url.URL) (filesystem.FileSystem, error) {
	conf := &Config{}
	_ = conf.Decode(uri.Query())

	return New(conf, "/"), nil

}
