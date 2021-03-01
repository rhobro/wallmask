package idx

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/rhobro/goutils/pkg/coll"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/wallmask/pkg/proxy"
	"net/http"
	"strings"
	"time"
)

func init() {
	src := "proxydb.net"
	base := "http://www.proxydb.net/?protocol=http&protocol=https&protocol=socks5&anonlvl=4"

	run := func() {
		rq, _ := http.NewRequest("GET", base, nil)
		rq.Header.Set("User-Agent", httputil.RandUA())
		rsp, err := httputil.RQUntil(http.DefaultClient, rq)
		if err != nil {
			proxyErr(src, err)
			return
		}
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		rsp.Body.Close()
		if err != nil {
			proxyErr(src, err)
			return
		}

		// Process
		page.Find("table > tbody > tr").Each(func(_ int, sl *goquery.Selection) {
			ip := sl.Find("td > a").Get(0).FirstChild.Data
			proto := strings.TrimSpace(sl.Find("td").Get(4).FirstChild.Data)

			if coll.ContainsStr([]string{"HTTP", "HTTPS", "SOCKS5"}, proto) {
				p, err := proxy.New(ip)
				if err != nil {
					return
				}
				p.Protocol = proxy.Protocol(strings.ToLower(proto))
				Add(p)
			}
		})
	}

	idxrs[src] = &idx{
		Period: 1 * time.Hour,
		run:    run,
	}
}
