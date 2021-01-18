package proxy

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
)

const (
	HTTP   Protocol = "http"
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
func New(raw string) (p *Proxy) {
	// Check if ip is valid and has port
	if strings.Count(raw, ".") == 3 && strings.Count(raw, ":") == 1 {
		spl := strings.Split(strings.TrimSpace(raw), ":")

		if len(spl) == 2 {
			port, err := strconv.Atoi(spl[1])
			if err != nil {
				log.Printf("invalid proxy raw string %s: %s", raw, err)
				return
			}

			if spl[0] != "" && port != 0 {
				p = &Proxy{
					IPv4: spl[0],
					Port: uint16(port),
				}
			}
		}
	}

	return
}
