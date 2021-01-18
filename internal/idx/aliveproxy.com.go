package idx

import (
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "aliveproxy.com"
	bases := map[proxy.Protocol]string{
		proxy.HTTP:   "http://www.aliveproxy.com/high-anonymity-proxy-list/",
		proxy.SOCKS5: "http://aliveproxy.com/socks5-list/",
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

		sl := page.Find("table.cm.or > tbody > tr")
		var ps []*proxy.Proxy
		sl.Each(func(row int, sl *goquery.Selection) {
			// Skip col headers
			if row == 0 {
				return
			}

			// check if valid cell and filter highly anonymous Proxies
			proxyType := strings.ToLower(sl.Find("td").Get(2).FirstChild.Data)
			if strings.Contains(proxyType, "high") {
				raw := strings.TrimSpace(sl.Find("td").Get(0).FirstChild.Data)
				p := proxy.New(raw)
				p.Protocol = sch
				ps = append(ps, p)
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
		Period: 5 * time.Minute,
		F:      run,
	}
}
