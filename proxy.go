package proxy

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type Proto string
type IPVersion string

const (
	HTTP   Proto = "http"
	SOCKS5 Proto = "socks5"
	SOCKS4 Proto = "socks4"

	IPv4 IPVersion = "4"
	IPv6 IPVersion = "6"

	payload string = "CONNECT %s HTTP/1.1\r\nHost: %s\r\n%sConnection: Keep-Alive\r\n\r\n" // target host, target host, auth header
)

var (
	ErrInvalidProxyString  = errors.New("invalid proxy string provided")
	ErrUnsupportedProtocol = errors.New("provided protocol not supported")
)

type Proxy struct {
	Proto    Proto
	Host     string
	Port     string
	Username string
	Password string
}

type Config struct {
	IPVersion IPVersion
	Proxy     Proxy
	Timeout   time.Duration
}

// proxy in format host:port[:username[:password]]
func Dial(proto Proto, proxy, host string) (net.Conn, error) {
	switch proto {
	case HTTP:
		var (
			proxyHost  string
			authHeader string
		)

		parts := strings.Split(proxy, ":")
		switch len(parts) {
		case 2:
			proxyHost = parts[0] + ":" + parts[1]
		case 3:
			proxyHost = parts[0] + ":" + parts[1]
			authHeader = "Proxy-Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte(parts[2]+":")) + "\r\n"
		case 4:
			proxyHost = parts[0] + ":" + parts[1]
			authHeader = "Proxy-Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte(parts[2]+":"+parts[3])) + "\r\n"
		default:
			return nil, ErrInvalidProxyString
		}

		conn, err := net.Dial("tcp", proxyHost)
		if err != nil {
			return nil, err
		}

		if _, err := conn.Write([]byte(fmt.Sprintf(payload, host, host, authHeader))); err != nil {
			conn.Close()
			return nil, err
		}

		var (
			buf          = make([]byte, 512)
			parsedString string
		)

		for !strings.HasSuffix(parsedString, "\r\n\r\n") {
			n, err := conn.Read(buf)
			if err != nil {
				conn.Close()
				return nil, err
			}

			parsedString += string(buf[:n])
		}

		lines := strings.Split(parsedString, "\r\n")
		if !strings.Contains(lines[0], " 200 ") {
			conn.Close()
			return nil, fmt.Errorf("proxy returned invalid status code, first line: %s\r\n", lines[0])
		}

		return conn, nil
	case SOCKS5:
		return nil, fmt.Errorf("temporary not supported")
	case SOCKS4:
		return nil, fmt.Errorf("temporary not supported")
	default:
		return nil, ErrUnsupportedProtocol
	}
}

func DialTimeout(proto Proto, proxy, host string, timeout time.Duration) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	switch proto {
	case HTTP:
		var (
			proxyHost  string
			authHeader string
		)

		parts := strings.Split(proxy, ":")
		switch len(parts) {
		case 2:
			proxyHost = parts[0] + ":" + parts[1]
		case 3:
			proxyHost = parts[0] + ":" + parts[1]
			authHeader = "Proxy-Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte(parts[2]+":")) + "\r\n"
		case 4:
			proxyHost = parts[0] + ":" + parts[1]
			authHeader = "Proxy-Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte(parts[2]+":"+parts[3])) + "\r\n"
		default:
			return nil, ErrInvalidProxyString
		}

		conn, err := net.DialTimeout("tcp", proxyHost, timeout)
		if err != nil {
			return nil, err
		}
		conn.SetReadDeadline(deadline)

		if _, err := conn.Write([]byte(fmt.Sprintf(payload, host, host, authHeader))); err != nil {
			conn.Close()
			return nil, err
		}

		var (
			buf          = make([]byte, 512)
			parsedString string
		)

		for !strings.HasSuffix(parsedString, "\r\n\r\n") {
			n, err := conn.Read(buf)
			if err != nil {
				conn.Close()
				return nil, err
			}

			parsedString += string(buf[:n])
		}

		lines := strings.Split(parsedString, "\r\n")
		if !strings.Contains(lines[0], " 200 ") {
			conn.Close()
			return nil, fmt.Errorf("proxy returned invalid status code, first line: %s\r\n", lines[0])
		}

		return conn, nil
	case SOCKS5:
		return nil, fmt.Errorf("temporary not supported")
	case SOCKS4:
		return nil, fmt.Errorf("temporary not supported")
	default:
		return nil, ErrUnsupportedProtocol
	}
}

func DialCustom() (net.Conn, error) {
	return nil, fmt.Errorf("temporary dont work")
}
