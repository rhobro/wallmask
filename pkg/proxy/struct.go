package proxy

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	HTTP   Protocol = "http"
	HTTPS  Protocol = "https"
	SOCKS5 Protocol = "socks5"
)

type Protocol string

type Proxy struct {
	Protocol Protocol
	IPv4     string
	Port     uint16
}

// fmt.Stringer
func (p *Proxy) String() string {
	return fmt.Sprintf("%s://%s:%d", p.Protocol, p.IPv4, p.Port)
}

func (p *Proxy) URL() (*url.URL, error) {
	return url.Parse(p.String())
}

// Parse addresses in format ip:port
func New(raw string) (p *Proxy, err error) {
	// Check if ip is valid and has port
	if strings.Count(raw, ".") == 3 && strings.Count(raw, ":") == 1 {
		spl := strings.Split(strings.TrimSpace(raw), ":")

		if len(spl) == 2 {
			port, err := strconv.Atoi(spl[1])
			if err != nil {
				return nil, err
			}

			if spl[0] != "" && port != 0 {
				p = &Proxy{
					IPv4: spl[0],
					Port: uint16(port),
				}
			} else {
				return nil, errors.New("does not have a valid IP or port is 0")
			}
		}
	} else {
		err = errors.New("doesn't contain correct format of perids and semicolons")
	}

	return
}
