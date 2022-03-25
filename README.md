# go-vfs
virtual file system for go

## Tips
This project is not finished yet. Do not use in production environment

## How to use

```golang
package main

import (
	"fmt"
	"github.com/lazychanger/filesystem"
	_ "github.com/lazychanger/filesystem/driver/memory"
	_ "github.com/lazychanger/filesystem/driver/os"
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

- [ ] go test cover 100%
- [ ] more filesystem driver. eg. s3, etcd