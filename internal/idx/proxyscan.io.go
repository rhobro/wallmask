package idx

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/wallmask/pkg/proxy"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func init() {
	src := "proxyscan.io"
	base := "https://www.proxyscan.io/Home/FilterResult"
	v := url.Values{}
	v.Add("SelectedAnonymity", "Elite")
	v.Add("sortPing", "false")
	v.Add("sortTime", "true")
	v.Add("sortUptime", "false")

	scrape := func(sch proxy.Protocol, v *url.Values) {
		// Request
		rq, _ := http.NewRequest("POST", base, strings.NewReader(v.Encode()))
		rq.Header.Set("User-Agent", httputil.RandUA())
		rsp, err := httputil.RQUntil(http.DefaultClient, rq)
		if err != nil {
			proxyErr(src, fmt.Errorf("rq for list page: %s", err))
			return
		}
		defer rsp.Body.Close()
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		if err != nil {
			proxyErr(src, fmt.Errorf("parse page HTML: %s", err))
			return
		}

		page.Find("tr").Each(func(i int, sl *goquery.Selection) {
			ip := sl.Find("th").Get(0).FirstChild.Data
			port, err := strconv.Atoi(sl.Find("td").Get(0).FirstChild.Data)
			if err != nil {
				return
			}

			Add(&proxy.Proxy{
				Protocol: sch,
				IPv4:     ip,
				Port:     uint16(port),
			})
		})
	}

	run := func() {
		v.Add("selectedType", "HTTP")
		v.Add("selectedType", "HTTPS")
		scrape(proxy.HTTP, &v)
		v.Set("selectedType", "SOCKS5")
		scrape(proxy.SOCKS5, &v)
	}

	idxrs[src] = &idx{
		Period: 10 * time.Minute,
		run:    run,
	}
}
