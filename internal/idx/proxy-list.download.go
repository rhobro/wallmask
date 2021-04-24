package idx

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/wallmask/pkg/wallmask"
	"net/http"
	"strconv"
	"time"
)

func init() {
	src := "proxy-list.download"
	bases := map[string]wallmask.Protocol{
		"https://www.proxy-list.download/HTTP": wallmask.HTTP,
		"https://www.proxy-list.download/HTTPS": wallmask.HTTPS,
		"https://www.proxy-list.download/SOCKS5": wallmask.SOCKS5,
	}

	scrape := func(sch wallmask.Protocol, base string) {
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

		page.Find("table#tbl > tbody > tr").Each(func(i int, sl *goquery.Selection) {
			ip := sl.Find("td").Get(0).FirstChild.Data
			port, err := strconv.Atoi(sl.Find("td").Get(1).FirstChild.Data)
			if err != nil {
				return
			}

			Add(&wallmask.Proxy{
				Proto: sch,
				IPv4:  ip,
				Port:  uint16(port),
			})
		})
	}

	run := func(bool) {
		for ur, sc := range bases {
			scrape(sc, ur)
		}
	}

	idxrs[src] = &idx{
		Period: time.Hour,
		run:    run,
	}
}
