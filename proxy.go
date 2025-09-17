package proxy

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

var (
	InvalidHostFormat     = errors.New("invalid host format")
	ProxyConnectionFailed = errors.New("proxy connection failed")
)

type TCPOption func(*net.TCPConn) error

// if true, data will be set as soon as possible
// if false, OS will use Nagle`s algorithm to optimize tcp
// default: true
func WithNoDelay() TCPOption {
	return func(c *net.TCPConn) error {
		return c.SetNoDelay(true)
	}
}

// sets the duration the connection needs to remain idle before TCP starts sending keepalive probes
func WithKeepAlive(d time.Duration) TCPOption {
	return func(c *net.TCPConn) error {
		c.SetKeepAlive(true)
		return c.SetKeepAlivePeriod(d)
	}
}

// sets the read and write deadlines associated with the connection
func WithDeadline(t time.Time) TCPOption {
	return func(c *net.TCPConn) error {
		return c.SetDeadline(t)
	}
}

// sets the read deadline associated with the connection
func WithReadDeadline(t time.Time) TCPOption {
	return func(c *net.TCPConn) error {
		return c.SetReadDeadline(t)
	}
}

// sets the write deadline associated with the connection
func WithWriteDeadline(t time.Time) TCPOption {
	return func(c *net.TCPConn) error {
		return c.SetWriteDeadline(t)
	}
}

// if true, will abort connection right after calling close()
// if false, will wait until all data sent (but wont block the program)
// default: false
// (use false if you dont know what is it)
func WithLinger(abort bool) TCPOption {
	return func(c *net.TCPConn) error {
		if abort {
			return c.SetLinger(0)
		}
		return c.SetLinger(-1)
	}
}

// sets the size of the OS recieve buffer associated with connection
func WithReadBuffer(bytes int) TCPOption {
	return func(c *net.TCPConn) error {
		return c.SetReadBuffer(bytes)
	}
}

// sets the size of the OS transmit buffer associated with connection
func WithWriteBuffer(bytes int) TCPOption {
	return func(c *net.TCPConn) error {
		return c.SetWriteBuffer(bytes)
	}
}

// network is tcp network, tcp/tcp4/tcp6
// proxy require "host:port" or "host:port:user:password"
// targetHost require target host details
func DialHTTP(network, proxy string, targetHost string, options ...TCPOption) (net.Conn, error) {
	var (
		raddr *net.TCPAddr
		err   error
		auth  string
	)
	parts := strings.SplitN(proxy, ":", 4)
	if len(parts) > 2 && len(parts) < 5 {
		var passwd string

		if len(parts) == 4 {
			passwd = parts[3]
		}

		raddr, err = net.ResolveTCPAddr(network, parts[0]+":"+parts[1])

		auth = "Proxy-Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte(parts[2]+":"+passwd)) + "\r\n"

	} else if len(parts) == 2 {
		raddr, err = net.ResolveTCPAddr(network, proxy)
	} else {
		return nil, InvalidHostFormat
	}

	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP(network, nil, raddr)
	if err != nil {
		return nil, err
	}

	payload := []byte(
		"CONNECT " + targetHost + " HTTP/1.1\r\n" +
			"Host: " + targetHost + "\r\n" +
			auth +
			"Connection: keep-alive\r\n" +
			"Proxy-Connection: keep-alive\r\n\r\n",
	)

	if _, err := conn.Write(payload); err != nil {
		return nil, err
	}

	r := bufio.NewReader(conn)

	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	if !strings.Contains(line, " 200 ") {
		return nil, fmt.Errorf("proxy returned bad status, status line: %s", line)
	}

	for line != "\r\n" {
		line, err = r.ReadString('\n')
		if err != nil {
			return nil, err
		}
	}

	for _, o := range options {
		if err := o(conn); err != nil {
			return nil, err
		}
	}
	return conn, nil
}
