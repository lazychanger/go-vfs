package os

import (
	"github.com/lazychanger/filesystem"
)

type Config struct {
	filesystem.Config

	Root string
}

func (conf *Config) Driver() string {
	return Driver
}

func (conf *Config) Path() string {
	return conf.Root
}
