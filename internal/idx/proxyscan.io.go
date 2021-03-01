package idx

import (
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
	base := "https://www.proxyscan.io/home/filterresult"
	v := url.Values{}
	v.Set("limit", "100")
	v.Set("SelectedAnonymity", "Elite")
	v.Set("sortPing", "false")
	v.Set("sortTime", "true")
	v.Set("sortUptime", "false")
	v.Add("selectedType", "HTTP")
	v.Add("selectedType", "HTTPS")
	v.Add("selectedType", "SOCKS5")
	v.Set("page", "0")

	run := func() {
		v.Set("_", strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
		// Request
		rq, _ := http.NewRequest("POST", base, strings.NewReader(v.Encode()))
		rq.Header.Set("User-Agent", httputil.RandUA())
		rsp, err := httputil.RQUntil(http.DefaultClient, rq)
		if err != nil {
			//proxyErr(src, err)
			return
		}
		defer rsp.Body.Close()
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		if err != nil {
			//proxyErr(src, err)
			return
		}

		page.Find("tr").Each(func(i int, sl *goquery.Selection) {
			ip := sl.Find("th").Get(0).FirstChild.Data
			port, err := strconv.Atoi(sl.Find("td").Get(0).FirstChild.Data)
			protos := strings.TrimSpace(sl.Find("td").Get(3).FirstChild.Data)
			if err != nil {
				return
			}

			p := &proxy.Proxy{
				IPv4: ip,
				Port: uint16(port),
			}
			if !strings.Contains(protos, ",") {
				switch protos {
				case "HTTP":
					p.Proto = proxy.HTTP
				case "HTTPS":
					p.Proto = proxy.HTTPS
				case "SOCKS5":
					p.Proto = proxy.SOCKS5
				}
			}
			Add(p)
		})
	}

	idxrs[src] = &idx{
		Period: 10 * time.Minute,
		run:    run,
	}
}
