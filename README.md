# inproc
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


