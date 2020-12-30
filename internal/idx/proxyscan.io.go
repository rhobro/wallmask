package idx

import (
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "proxyscan.io"
	run := func() {
		base := "https://www.proxyscan.io/Home/FilterResult"
		v := url.Values{}
		v.Add("selectedType", "HTTP")
		v.Add("selectedType", "HTTPS")
		v.Add("SelectedAnonymity", "Elite")
		refreshDuration := 10 * time.Minute
		t := time.NewTicker(refreshDuration)

		for {
			// Request
			rq, _ := http.NewRequest("POST", base, strings.NewReader(v.Encode()))
			rq.Header.Set("User-Agent", httputil.RandUA())
			rsp, err := httputil.RQUntil(http.DefaultClient, rq)
			if err != nil {
				proxyErr(src, fmt.Errorf("rq for list page: %s", err))
				continue
			}
			bd, err := ioutil.ReadAll(rsp.Body)
			// Add html tags to allow parser to work
			html := "<!DOCTYPE html><html><head></head><body><table>" + string(bd) + "</table></body></html>"
			page, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				proxyErr(src, fmt.Errorf("parse page HTML: %s", err))
				continue
			}
			rsp.Body.Close()

			page.Find("tr").Each(func(_ int, sl *goquery.Selection) {
				ip := sl.Find("th").Get(0).FirstChild.Data
				port, err := strconv.Atoi(sl.Find("td").Get(0).FirstChild.Data)
				if err != nil {
					return
				}

				Add(&proxy.Proxy{
					IPv4: ip,
					Port: uint16(port),
				})
			})

			<-t.C
		}
	}

	addFuncs[src] = run
}
