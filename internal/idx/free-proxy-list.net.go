package idx

import (
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"time"
	"wallmask/pkg/proxy"
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
			proxyErr(src, fmt.Errorf("rq for list page: %s", err))
			return
		}
		defer rsp.Body.Close()
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		if err != nil {
			proxyErr(src, fmt.Errorf("parse page HTML: %s", err))
			return
		}

		rawList := strings.Split(page.Find("div.modal-body > textarea").Get(0).FirstChild.Data, "\n\n")[1]
		for _, line := range strings.Split(strings.TrimSpace(rawList), "\n") {
			p := proxy.New(line)
			p.Protocol = proxy.HTTP
			Add(p)
		}
	}

	idxrs[src] = &idx{
		Period: 30 * time.Minute,
		run:    run,
	}
}
