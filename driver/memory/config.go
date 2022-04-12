package memory

import (
	"github.com/lazychanger/go-vfs"
	"net/url"
	"strconv"
)

type Config struct {
	filesystem.Config

	MaxSize int64
}

func (conf *Config) Driver() string {
	return Driver
}

// Encode the options to url.Values
func (conf *Config) Encode() url.Values {
	return url.Values{
		"maxsize": []string{strconv.FormatInt(conf.MaxSize, 10)},
	}
}

// Decode the url.Values to options
func (conf *Config) Decode(query url.Values) error {

	conf.MaxSize, _ = strconv.ParseInt(query.Get("maxsize"), 10, 64)

	return nil
}
