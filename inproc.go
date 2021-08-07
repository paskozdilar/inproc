// Copyright (c) 2021 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// Package inproc implements an in-process connection.
package inproc

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const network = "inproc"

type addr struct {
	network string
	address string
}

func (a addr) Network() string {
	return a.network
}

func (a addr) String() string {
	return a.address
}

var inprocs struct {
	locker    sync.RWMutex
	listeners map[addr]*listener
}

func init() {
	inprocs.listeners = make(map[addr]*listener)
}

type conn struct {
	r     io.ReadCloser
	w     io.WriteCloser
	laddr addr
	raddr addr
}

// Read reads data from the connection.
func (c *conn) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

// Write writes data to the connection.
func (c *conn) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

// Close closes the connection.
func (c *conn) Close() (err error) {
	if c.w != nil {
		err = c.w.Close()
	}
	if c.r != nil {
		err = c.r.Close()
	}
	return
}

// LocalAddr returns the local network address.
func (c *conn) LocalAddr() net.Addr {
	return c.laddr
}

// RemoteAddr returns the remote network address.
func (c *conn) RemoteAddr() net.Addr {
	return c.raddr
}

// SetDeadline implements the Conn SetDeadline method.
func (c *conn) SetDeadline(t time.Time) error {
	return errors.New("not supported")
}

// SetReadDeadline implements the Conn SetReadDeadline method.
func (c *conn) SetReadDeadline(t time.Time) error {
	return errors.New("not supported")
}

// SetWriteDeadline implements the Conn SetWriteDeadline method.
func (c *conn) SetWriteDeadline(t time.Time) error {
	return errors.New("not supported")
}

// Dial connects to an address.
func Dial(address string) (net.Conn, error) {
	raddr := addr{network: network, address: address}
	var accepter *accepter
	r, w := io.Pipe()
	conn := &conn{w: w, laddr: raddr}
	inprocs.locker.RLock()
	l, ok := inprocs.listeners[raddr]
	if !ok {
		inprocs.locker.RUnlock()
		return nil, errors.New("connection refused")
	}
	inprocs.locker.RUnlock()
	l.locker.Lock()
	for {
		if len(l.accepters) > 0 {
			accepter = l.accepters[len(l.accepters)-1]
			l.accepters = l.accepters[:len(l.accepters)-1]
			break
		}
		l.cond.Wait()
	}
	l.locker.Unlock()
	conn.r = accepter.reader
	conn.raddr = conn.laddr
	accepter.conn.r = r
	accepter.conn.raddr = conn.laddr
	close(accepter.done)
	return conn, nil
}

type listener struct {
	laddr     addr
	cond      sync.Cond
	locker    sync.Mutex
	accepters []*accepter
	done      chan struct{}
	closed    int32
}

type accepter struct {
	*conn
	reader io.ReadCloser
	done   chan struct{}
}

// Listen announces on the local address.
func Listen(address string) (net.Listener, error) {
	laddr := addr{network: network, address: address}
	l := &listener{laddr: laddr, done: make(chan struct{})}
	l.cond.L = &l.locker
	inprocs.locker.Lock()
	if _, ok := inprocs.listeners[l.laddr]; ok {
		inprocs.locker.Unlock()
		return nil, errors.New("address already in use")
	}
	inprocs.listeners[l.laddr] = l
	inprocs.locker.Unlock()
	return l, nil
}

// Accept waits for and returns the next connection to the listener.
func (l *listener) Accept() (net.Conn, error) {
	r, w := io.Pipe()
	accepter := &accepter{conn: &conn{w: w, laddr: l.laddr}, reader: r}
	accepter.done = make(chan struct{})
	l.locker.Lock()
	l.accepters = append(l.accepters, accepter)
	l.locker.Unlock()
	l.cond.Broadcast()
	select {
	case <-accepter.done:
		return accepter.conn, nil
	case <-l.done:
		return nil, errors.New("use of closed network connection")
	}
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *listener) Close() error {
	if atomic.CompareAndSwapInt32(&l.closed, 0, 1) {
		close(l.done)
	}
	inprocs.locker.Lock()
	delete(inprocs.listeners, l.laddr)
	inprocs.locker.Unlock()
	l.locker.Lock()
	accepters := l.accepters
	l.accepters = nil
	l.locker.Unlock()
	l.cond.Broadcast()
	for _, accepter := range accepters {
		accepter.Close()
	}
	return nil
}

// Addr returns the listener's network address.
func (l *listener) Addr() net.Addr {
	return l.laddr
}
