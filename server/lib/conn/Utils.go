package conn

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const PREFIX_UNIX = "unix:"

func CreateListener(addr string) (net.Listener, error) {
	if strings.HasPrefix(addr, PREFIX_UNIX) {
		unixAddr, err := GetUnixAddress(addr)
		if err != nil {
			return nil, err
		}
		if s, err := os.Stat(unixAddr); !os.IsNotExist(err) {
			err = os.Remove(s.Name())
			if err != nil {
				return nil, err
			}
		}
		l, err := net.Listen("unix", unixAddr)
		if err != nil {
			return nil, err
		}
		return l, os.Chmod(unixAddr, 0666)
	}
	return net.Listen("tcp4", addr)
}

func CreateURLConnection(u *url.URL) (net.Conn, error) {
	return CreateConnection(GetAddress(u))
}

func CreateConnection(addr string) (net.Conn, error) {
	if strings.HasPrefix(addr, PREFIX_UNIX) {
		unixAddr, err := GetUnixAddress(addr)
		if err != nil {
			return nil, err
		}
		return net.Dial("unix", unixAddr)
	}
	return net.Dial("tcp4", addr)
}

func GetAddress(u *url.URL) string {
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if u.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func GetUnixAddress(addr string) (string, error) {
	return filepath.Abs(strings.ReplaceAll(addr, PREFIX_UNIX, ""))
}
