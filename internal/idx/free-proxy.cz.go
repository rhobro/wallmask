// +build ignore

package idx

import (
	"encoding/base64"
	"github.com/PuerkitoBio/goquery"
	"github.com/rhobro/goutils/pkg/httputil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func init() {
	src := "free-proxy.cz"
	urls := []string{
		"http://free-proxy.cz/en/proxylist/country/all/http/ping/level1",
		"http://free-proxy.cz/en/proxylist/country/all/https/ping/level1",
	}
	extractB64RGX, _ := regexp.Compile("\"(.*)\"")

	run := func(bool) {
		for _, base := range urls {
			// Request
			rq, _ := http.NewRequest("GET", base, nil)
			rq.Header.Set("User-Agent", httputil.RandUA())
			rsp, err := httputil.RQUntil(http.DefaultClient, rq)
			if err != nil {
				proxyErr(src, err)
				continue
			}
			page, err := goquery.NewDocumentFromReader(rsp.Body)
			if err != nil {
				proxyErr(src, err)
				continue
			}
			rsp.Body.Close()
			page.Find("table#proxy_list > tbody > tr").Each(func(i int, sl *goquery.Selection) {
				if node := sl.Find("td > script"); node.Length() > 0 {
					raw := string(extractB64RGX.Find([]byte(node.Get(0).FirstChild.Data)))
					raw = raw[1 : len(raw)-1]
					ipBytes, err := base64.StdEncoding.DecodeString(raw)
					if err != nil {
						proxyErr(src, err)
						return
					}

					port, err := strconv.Atoi(sl.Find("td > span").Get(0).FirstChild.Data)
					if err != nil {
						proxyErr(src, err)
						return
					}

					Add(&proxy.Proxy{
						IPv4: string(ipBytes),
						Port: uint16(port),
					})
				}
			})
			AddBatch(ps)
		}
	}

	idxrs[src] = &idx{
		Period: time.Hour,
		run:    run,
	}
}
