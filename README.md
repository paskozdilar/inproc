# inproc

[![PkgGoDev](https://pkg.go.dev/badge/github.com/hslam/inproc)](https://pkg.go.dev/github.com/hslam/inproc)
[![Build Status](https://github.com/hslam/inproc/workflows/build/badge.svg)](https://github.com/hslam/inproc/actions)
[![codecov](https://codecov.io/gh/hslam/inproc/branch/master/graph/badge.svg)](https://codecov.io/gh/hslam/inproc)
[![Go Report Card](https://goreportcard.com/badge/github.com/hslam/inproc)](https://goreportcard.com/report/github.com/hslam/inproc)
[![LICENSE](https://img.shields.io/github/license/hslam/inproc.svg?style=flat-square)](https://github.com/hslam/inproc/blob/master/LICENSE)

Package inproc implements an in-process connection.

## Features

* In process
* Compatible with the net.Conn interface.

## Get started

### Install
```
go get github.com/hslam/inproc
```

### Import
```
import "github.com/hslam/inproc"
```

### Usage
#### Example
```go
package main

import (
	"fmt"
	"github.com/hslam/inproc"
	"net"
)

func main() {
	address := ":8080"
	l, err := inproc.Listen(address)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				buf := make([]byte, 1024)
				for {
					n, err := conn.Read(buf)
					if err != nil {
						break
					}
					conn.Write(buf[:n])
				}
				conn.Close()
			}(conn)
		}
	}()
	conn, err := inproc.Dial(address)
	if err != nil {
		panic(err)
	}
	msg := "Hello World"
	if _, err := conn.Write([]byte(msg)); err != nil {
		panic(err)
	}
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(buf[:n]))
	conn.Close()
}
```

### License
This package is licensed under a MIT license (Copyright (c) 2021 Meng Huang)


### Author
inproc was written by Meng Huang.


