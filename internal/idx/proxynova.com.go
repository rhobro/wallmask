package idx

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/wallmask/pkg/proxy"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	src := "proxynova.com"
	base := "https://www.proxynova.com/proxy-server-list/elite-proxies/"

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

		page.Find("table.table > tbody > tr").Each(func(i int, sl *goquery.Selection) {
			if sl.Find("td > abbr > script").Length() > 0 {
				ip := sl.Find("td > abbr > script").Get(0).FirstChild.Data
				ip = ip[strings.Index(ip, "'")+1:]
				ip = ip[:strings.Index(ip, "'")]

				port, err := strconv.Atoi(sl.Find("td").Get(1).FirstChild.Data)
				if err != nil {
					return
				}

				Add(&proxy.Proxy{
					IPv4: ip,
					Port: uint16(port),
				})
			}
		})
	}

	idxrs[src] = &idx{
		Period: time.Minute,
		run:    run,
	}
}
