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
	src := "hidemy.name"
	base := "https://hidemy.name/en/proxy-list/?type=hs5"

	run := func(bool) {
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

		page.Find("div.table_block > table > tbody > tr").Each(func(i int, sl *goquery.Selection) {
			ip := sl.Find("td").Get(0).FirstChild.Data
			port, err := strconv.Atoi(sl.Find("td").Get(1).FirstChild.Data)
			if err != nil {
				return
			}
			Add(&wallmask.Proxy{
				IPv4: ip,
				Port: uint16(port),
			})
		})
	}

	idxrs[src] = &idx{
		Period: 5 * time.Minute,
		run:    run,
	}
}
