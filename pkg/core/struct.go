package core

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

type Proxy struct {
	IP   string
	Port uint16
}

// fmt.Stringer
func (p *Proxy) String() string {
	return fmt.Sprintf("http://%s:%d", p.IP, p.Port)
}

func (p *Proxy) URL() (*url.URL, error) {
	return url.Parse(p.String())
}

func New(raw string) (p *Proxy) {
	// Check if ip is valid and has port
	if strings.Count(raw, ".") == 3 && strings.Count(raw, ":") == 1 {
		spl := strings.Split(strings.TrimSpace(raw), ":")

		if len(spl) == 2 {
			port, err := strconv.ParseUint(spl[1], 10, 16)
			if err != nil {
				return
			}

			if spl[0] != "" && port != 0 {
				p = &Proxy{
					IP:   spl[0],
					Port: uint16(port),
				}
			}
		}
	}

	return
}
