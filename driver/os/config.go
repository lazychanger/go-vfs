package os

import (
	"github.com/lazychanger/filesystem"
	"net/url"
)

type Config struct {
	filesystem.Config

	Root string
}

// Encode the options to url.Values
func (conf *Config) Encode() url.Values {
	return url.Values{}
}

// Decode the url.Values to options
func (conf *Config) Decode(query url.Values) error {
	return nil
}

func (conf *Config) Driver() string {
	return Driver
}

func (conf *Config) Path() string {
	return conf.Root
}
