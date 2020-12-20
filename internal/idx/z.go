package idx

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"wallmask/pkg/core"
)

var addFuncs = make(map[string]func())

func init() {
	// launch indexers
	for src, f := range addFuncs {
		log.Printf("{proxy} Indexing proxies from %s", src)
		//go f() TODO
		_ = f
	}
	log.Println("{proxy} Initialized")

	// TODO test proxies
}

func Add(p *core.Proxy) {
	if p != nil {
		// Add if positive test
		if test(p) {
			// Add to database TODO
		}
	}
}

const testTimeout = 1 * time.Second // TODO use to remove slow proxies

func test(p *core.Proxy) (successful bool) {
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
		rsp, err := cli.Get("https://bytesimal.github.io/test")

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
