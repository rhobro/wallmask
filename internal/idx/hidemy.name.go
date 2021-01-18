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
	src := "hidemy.name"
	base := "https://hidemy.name/en/proxy-list/?type=hs5"

	run := func() {
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

		sl := page.Find("div.table_block > table > tbody > tr")
		ps := make([]*proxy.Proxy, sl.Length(), sl.Length())
		sl.Each(func(i int, sl *goquery.Selection) {
			ip := sl.Find("td").Get(0).FirstChild.Data
			port, err := strconv.Atoi(sl.Find("td").Get(1).FirstChild.Data)
			if err != nil {
				return
			}
			ps[i] = &proxy.Proxy{
				IPv4: ip,
				Port: uint16(port),
			}
		})
		AddBatch(ps)
	}

	idxrs[src] = &idx{
		Period: 5 * time.Minute,
		F:      run,
	}
}
