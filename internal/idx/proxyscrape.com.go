package idx

import (
	"bufio"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/wallmask/pkg/wallmask"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func init() {
	src := "proxyscrape.com"
	base := "https://api.proxyscrape.com/v2/?"

	scrape := func(sch wallmask.Protocol, v *url.Values) {
		// Request
		rq, _ := http.NewRequest("GET", base+v.Encode(), nil)
		rq.Header.Set("User-Agent", httputil.RandUA())
		rsp, err := httputil.RQUntil(http.DefaultClient, rq)
		if err != nil {
			proxyErr(src, err)
			return
		}
		rd := bufio.NewReader(rsp.Body)

		for {
			line, err := rd.ReadString('\n')
			// Check for EOF or error
			if err != nil {
				if err != io.EOF {
					proxyErr(src, err)
				}
				break
			}
			line = strings.TrimSpace(line)

			// add after parsing string
			p, err := wallmask.New(line)
			if err == nil {
				Add(p)
			}
		}
		rsp.Body.Close()
	}

	run := func(bool) {
		v := url.Values{}
		v.Set("request", "displayproxies")
		v.Set("protocol", "http")
		v.Set("timeout", "10000")
		v.Set("country", "all")
		v.Set("ssl", "all")
		v.Set("anonymity", "elite")
		scrape(wallmask.HTTP, &v)
		v = url.Values{}
		v.Set("request", "displayproxies")
		v.Set("protocol", "socks5")
		v.Set("timeout", "10000")
		v.Set("country", "all")
		scrape(wallmask.SOCKS5, &v)
	}

	idxrs[src] = &idx{
		Period: 5 * time.Minute,
		run:    run,
	}
}
