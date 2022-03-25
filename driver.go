package filesystem

import (
	"errors"
	"fmt"
	"net/url"
)

var (
	errNotSupported = "not supported driver `%s`"
)

var drivers = make(map[string]Driver)

// RegisterDriver register FileSystem
func RegisterDriver(driver string, d Driver) {
	drivers[driver] = d
}

type Driver interface {
	Open(uri *url.URL) (FileSystem, error)
}

// Open driver, return fs.FS and error
// use uri.Scheme to get driver
// use uri.User to get auth
// use uri.Host to get host
// use uri.Path to get root path
// use uri.RawQuery to get setting
// eg. os:///tmp/filesystem?a=1&b=2
//     memory:///?maxsize=10240000
// 	   s3://user:pass@host/bucket?region=us-east-1
//	   ftp://user:pass@host/path?passive=true
// 	   sftp://user:pass@host/path?passive=true
func Open(dns string) (FileSystem, error) {
	uri, err := url.Parse(dns)
	if err != nil {
		return nil, err
	}
	if f, ok := drivers[uri.Scheme]; ok {
		return f.Open(uri)
	}

	return nil, errors.New(fmt.Sprintf(errNotSupported, uri.Scheme))
}
