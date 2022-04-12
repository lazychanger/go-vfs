# go-vfs

[![codecov](https://codecov.io/gh/lazychanger/go-vfs/branch/master/graph/badge.svg?token=GUBD053ODP)](https://codecov.io/gh/lazychanger/go-vfs)
[![Go](https://img.shields.io/badge/go-1.18-blue.svg)](https://golang.org/doc/install)
[![MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)



virtual file system for go


## How to use

```golang
package main

import (
	"fmt"
	"github.com/lazychanger/go-vfs"
	_ "github.com/lazychanger/go-vfs/driver/memory"
	_ "github.com/lazychanger/go-vfs/driver/os"
	"io"
	"log"
	"os"
)

func main() {
	memvfs, _ := filesystem.Open("memory:///?maxsize=1000000")
	
	// more 

	wd, _ := os.Getwd()
	
	osvfs, _ := filesystem.Open(fmt.Sprintf("os://%s/runtime", wd))

	// more

}

```

## Features

- [ ] more filesystem driver. eg. s3, etcd