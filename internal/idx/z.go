package idx

import (
	"crypto/tls"
	"fmt"
	"github.com/Bytesimal/goutils/pkg/util"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
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

func Rand() func(r *http.Request) (*url.URL, error) {
	return func(r *http.Request) (*url.URL, error) {
		// Wait until at least 1 idx in map
		for len(Proxies) == 0 {
			continue
		}

		mtx.RLock()
		defer mtx.RUnlock()
		n := util.Rand.Intn(len(Proxies))

		var i int
		for _, p := range Proxies {
			if i == n {
				log.Println(p.String())
				return p.URL()
			}
			i++
		}
		return nil, nil
	}
}

var mtx sync.RWMutex
var Proxies = make(map[string]*Proxy)

func Add(p *Proxy) {
	if p != nil {
		// Add if positive test
		go func() {
			if test(p) {
				mtx.Lock()
				defer mtx.Unlock()
				Proxies[p.String()] = p
			}
		}()
	}
}

// Used to set a max number of concurrent connections
var semaphore = make(chan struct{}, 50)

const testTimeout = 1 * time.Second // TODO use to remove slow proxies

func test(p *Proxy) (successful bool) {
	u, err := p.URL()
	if err == nil {
		cli := http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(u),
				TLSClientConfig: &tls.Config{},
			},
			Timeout: testTimeout,
		}

		// test request to https://bytesimal.github.io/test
		semaphore <- struct{}{}
		rsp, err := cli.Get("https://bytesimal.github.io/test")
		<-semaphore

		if err == nil {
			defer rsp.Body.Close()
			bd, err := ioutil.ReadAll(rsp.Body)
			if err == nil {
				// Check response contains "TEST PAGE"
				if strings.Contains(string(bd), "TEST PAGE") {
					successful = true
				}
			}
		}
	}
	return
}

var addFuncs = make(map[string]func())

func init() {
	// launch indexers
	for src, f := range addFuncs {
		log.Printf("{proxy} Indexing proxies from %s", src)
		go f()
	}
	log.Println("{proxy} Initialized")
}
