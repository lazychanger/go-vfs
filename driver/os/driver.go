package os

import (
	"github.com/lazychanger/filesystem"
	"net/url"
)

const Driver = "os"

func init() {
	filesystem.RegisterDriver(Driver, &osDriver{})
}

type osDriver struct {
}

// Open opens a file using the given path
func (o *osDriver) Open(uri *url.URL) (filesystem.FileSystem, error) {
	return New(&Config{
		Root: uri.Path,
	})
}
