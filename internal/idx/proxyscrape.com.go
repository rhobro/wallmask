package idx

import (
	"bufio"
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "proxyscrape.com"
	base := "https://api.proxyscrape.com/v2/?"

	scrape := func(sch proxy.Protocol, v *url.Values) {
		// Request
		rq, _ := http.NewRequest("GET", base+v.Encode(), nil)
		rq.Header.Set("User-Agent", httputil.RandUA())
		rsp, err := httputil.RQUntil(http.DefaultClient, rq)
		if err != nil {
			proxyErr(src, fmt.Errorf("rq proxy list: %s", err))
			return
		}
		rd := bufio.NewReader(rsp.Body)

		for {
			line, err := rd.ReadString('\n')
			// Check for EOF or error
			if err != nil {
				if err != io.EOF {
					proxyErr(src, fmt.Errorf("reading list: %s", err))
				}
				break
			}
			line = strings.TrimSpace(line)

			// add after parsing string
			p := proxy.New(line)
			p.Protocol = sch
			Add(p)
		}
		rsp.Body.Close()
	}

	run := func() {
		v := url.Values{}
		v.Set("request", "displayproxies")
		v.Set("protocol", "http")
		v.Set("timeout", "10000")
		v.Set("country", "all")
		v.Set("ssl", "all")
		v.Set("anonymity", "elite")
		scrape(proxy.HTTP, &v)
		v = url.Values{}
		v.Set("request", "displayproxies")
		v.Set("protocol", "socks5")
		v.Set("timeout", "10000")
		v.Set("country", "all")
		scrape(proxy.SOCKS5, &v)
	}

	idxrs[src] = &idx{
		Period: 5 * time.Minute,
		run:    run,
	}
}
