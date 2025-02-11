// Copyright (c) 2021 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

package inproc

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestINPROC(t *testing.T) {
	address := ":9999"
	if _, err := Dial(address); err == nil {
		t.Error(err)
	}
	l, err := Listen(address)
	if err != nil {
		t.Error(err)
	}
	if _, err := Listen(address); err == nil {
		t.Error()
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func(conn net.Conn) {
				defer wg.Done()
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
	conn, err := Dial(address)
	if err != nil {
		t.Error(err)
	}
	conn.SetWriteDeadline(time.Now().Add(time.Second))
	conn.SetReadDeadline(time.Now().Add(time.Second))
	conn.SetDeadline(time.Now().Add(time.Second))
	conn.LocalAddr()
	raddr := conn.RemoteAddr()
	raddr.Network()
	raddr.String()

	str := "Hello World"
	conn.Write([]byte(str))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Error(err)
	}
	if string(buf[:n]) != str {
		t.Errorf("error %s != %s", string(buf[:n]), str)
	}
	time.Sleep(time.Millisecond)
	conn.Close()
	l.Addr()
	l.Close()
	wg.Wait()
}

func TestDeadline(t *testing.T) {
	address := ":9999"
	l, err := Listen(address)
	if err != nil {
		t.Error(err)
		return
	}
	defer l.Close()
	go func() {
		for i := 0; i < 2; i++ {
			conn, err := l.Accept()
			if err != nil {
				t.Error(err)
				return
			}
			if _, err := conn.Read(make([]byte, 1)); err != nil {
				t.Error(err)
				return
			}
			if _, err := conn.Write([]byte{0}); err != nil {
				t.Error(err)
				return
			}
			conn.Close()
		}
	}()
	conn, err := Dial(address)
	if err != nil {
		t.Error(err)
		return
	}

	// Test SetReadDeadline
	conn.SetReadDeadline(time.Now().Add(1))
	n, _ := conn.Read(make([]byte, 1))
	if n != 0 {
		t.Errorf("read deadline error %d != %d", n, 0)
		return
	}
	// Write value to progress listener status
	if _, err := conn.Write([]byte{0}); err != nil {
		t.Error(err)
		return
	}
	// Reset read deadline
	conn.SetReadDeadline(time.Time{})

	// Test SetWriteDeadline
	conn.SetWriteDeadline(time.Now().Add(1))
	if _, err := conn.Write([]byte{0}); err == nil {
		t.Error(err)
		return
	}
	n, _ = conn.Write([]byte{0})
	if n != 0 {
		t.Errorf("write deadline error %d != %d", n, 0)
		return
	}
	// Read value to progress listener status, then close connection
	if _, err := conn.Read(make([]byte, 1)); err != nil {
		t.Error(err)
		return
	}
	// Reset write deadline
	conn.SetWriteDeadline(time.Time{})
	conn.Close()

	// Test SetDeadline
	conn, err = Dial(address)
	if err != nil {
		t.Error(err)
		return
	}
	conn.SetDeadline(time.Now().Add(1))
	// Test read
	n, _ = conn.Read(make([]byte, 1))
	if n != 0 {
		t.Errorf("deadline error %d != %d", n, 0)
		return
	}
	// Reset deadline
	conn.SetDeadline(time.Time{})

	// Write value to progress listener status
	if _, err := conn.Write([]byte{0}); err != nil {
		t.Error(err)
		return
	}
	conn.SetDeadline(time.Now().Add(1))
	// Test write
	n, _ = conn.Write([]byte{0})
	if n != 0 {
		t.Errorf("deadline error %d != %d", n, 0)
		return
	}
	// Read value to progress listener status, then close connection
	_, err = conn.Read(make([]byte, 1))
	if err == nil {
		t.Error(err)
		return
	}
	conn.Close()
}
