package memory

import (
	"github.com/lazychanger/filesystem"
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
	if err := conf.Decode(uri.Query()); err != nil {
		return nil, err
	}

	return New(conf, "/"), nil

}
