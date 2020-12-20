package idx

import (
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"time"
)

func init() {
	run := func() {
		base := "http://www.aliveproxy.com/high-anonymity-proxy-list/"
		refreshDuration := 5 * time.Minute
		t := time.NewTicker(refreshDuration)

		for {
			// Request
			rq, _ := http.NewRequest("GET", base, nil)
			rq.Header.Set("User-Agent", httputil.RandUA())
			rsp, err := httputil.RQUntil(http.DefaultClient, rq)
			if err != nil {
				continue
			}
			page, err := goquery.NewDocumentFromReader(rsp.Body)
			if err != nil {
				continue
			}

			page.Find("table.cm.or > tbody > tr").Each(func(row int, sl *goquery.Selection) {
				// Skip col headers
				if row == 0 {
					return
				}

				// check if valid cell and filter highly anonymous proxies
				proxyType := strings.ToLower(sl.Find("td").Get(2).FirstChild.Data)
				if strings.Contains(proxyType, "high") {
					raw := strings.TrimSpace(sl.Find("td").Get(0).FirstChild.Data)
					Add(New(raw))
				}
			})

			<-t.C
		}
	}

	src := "aliveproxy.com"
	addFuncs[src] = run
}
