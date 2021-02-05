package idx

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/wallmask/pkg/proxy"
	"net/http"
	"strings"
	"time"
)

func init() {
	src := "free-proxy-list.net"
	base := "https://free-proxy-list.net/"

	run := func() {
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

		rawList := strings.Split(page.Find("div.modal-body > textarea").Get(0).FirstChild.Data, "\n\n")[1]
		for _, line := range strings.Split(strings.TrimSpace(rawList), "\n") {
			p, err := proxy.New(line)
			if err == nil {
				Add(p)
			}
		}
	}

	idxrs[src] = &idx{
		Period: 30 * time.Minute,
		run:    run,
	}
}
