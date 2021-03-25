package idx

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/wallmask/pkg/proxy"
	"net/http"
	"strings"
	"time"
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
			proxyErr(src, err)
			return
		}
		defer rsp.Body.Close()
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		if err != nil {
			proxyErr(src, err)
			return
		}

		page.Find("table.cm.or > tbody > tr").Each(func(row int, sl *goquery.Selection) {
			// Skip col headers
			if row == 0 {
				return
			}

			// check if valid cell and filter highly anonymous Proxies
			proxyType := strings.ToLower(sl.Find("td").Get(2).FirstChild.Data)
			if strings.Contains(proxyType, "high") {
				raw := strings.TrimSpace(sl.Find("td").Get(0).FirstChild.Data)
				p, err := proxy.New(raw)
				if err == nil {
					p.Proto = sch
					Add(p)
				}
			}
		})
	}

	run := func(bool) {
		for sc, ur := range bases {
			scrape(sc, ur)
		}
	}

	idxrs[src] = &idx{
		Period: 5 * time.Minute,
		run:    run,
	}
}
