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
		bd, err := ioutil.ReadAll(rsp.Body)
		// Add html tags to allow parser to work
		html := "<!DOCTYPE html><html><head></head><body><table>" + string(bd) + "</table></body></html>"
		page, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			proxyErr(src, fmt.Errorf("parse page HTML: %s", err))
			return
		}
		rsp.Body.Close()

		sl := page.Find("tr")
		ps := make([]*proxy.Proxy, sl.Length(), sl.Length())
		sl.Each(func(i int, sl *goquery.Selection) {
			ip := sl.Find("th").Get(0).FirstChild.Data
			port, err := strconv.Atoi(sl.Find("td").Get(0).FirstChild.Data)
			if err != nil {
				return
			}

			ps[i] = &proxy.Proxy{
				Protocol: sch,
				IPv4:     ip,
				Port:     uint16(port),
			}
		})
		AddBatch(ps)
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
		F:      run,
	}
}
