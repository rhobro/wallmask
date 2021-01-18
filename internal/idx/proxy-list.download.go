package idx

import (
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strconv"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "proxy-list.download"
	bases := map[proxy.Protocol]string{
		proxy.HTTP:   "https://www.proxy-list.download/HTTP",
		proxy.SOCKS5: "https://www.proxy-list.download/SOCKS5",
	}

	scrape := func(sch proxy.Protocol, base string) {
		// Request
		rq, _ := http.NewRequest("GET", base, nil)
		rq.Header.Set("User-Agent", httputil.RandUA())
		rsp, err := httputil.RQUntil(http.DefaultClient, rq)
		if err != nil {
			proxyErr(src, fmt.Errorf("rq for list page: %s", err))
			return
		}
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		if err != nil {
			proxyErr(src, fmt.Errorf("parse page HTML: %s", err))
			return
		}
		rsp.Body.Close()

		sl := page.Find("table#tbl > tbody > tr")
		ps := make([]*proxy.Proxy, sl.Length(), sl.Length())
		sl.Each(func(i int, sl *goquery.Selection) {
			ip := sl.Find("td").Get(0).FirstChild.Data
			port, err := strconv.Atoi(sl.Find("td").Get(1).FirstChild.Data)
			if err != nil {
				return
			}

			ps[i] = &proxy.Proxy{
				Protocol: sch,
				IPv4:     ip,
				Port:     uint16(port),
			}
		})
		AddBatch(ps)
	}

	run := func() {
		for sc, ur := range bases {
			scrape(sc, ur)
		}
	}

	idxrs[src] = &idx{
		Period: time.Hour,
		F:      run,
	}
}
