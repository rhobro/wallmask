package idx

import (
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "proxynova.com"
	run := func() {
		base := "https://www.proxynova.com/proxy-server-list/elite-proxies/"
		refreshDuration := time.Minute
		t := time.NewTicker(refreshDuration)

		for {
			// Request
			rq, _ := http.NewRequest("GET", base, nil)
			rq.Header.Set("User-Agent", httputil.RandUA())
			rsp, err := httputil.RQUntil(http.DefaultClient, rq)
			if err != nil {
				proxyErr(src, fmt.Errorf("rq for list page: %s", err))
				continue
			}
			page, err := goquery.NewDocumentFromReader(rsp.Body)
			if err != nil {
				proxyErr(src, fmt.Errorf("parse page HTML: %s", err))
				continue
			}
			rsp.Body.Close()

			page.Find("table.table > tbody > tr").Each(func(_ int, sl *goquery.Selection) {
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

			<-t.C
		}
	}

	addFuncs[src] = run
}
