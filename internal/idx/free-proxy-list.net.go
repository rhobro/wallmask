package idx

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "free-proxy-list.net"
	run := func() {
		base := "https://free-proxy-list.net/"
		refreshDuration := 30 * time.Minute
		t := time.NewTicker(refreshDuration)

		for {
			// Request
			rq, _ := http.NewRequest("GET", base, nil)
			//rq.Header.Set("User-Agent", httputil.RandUA())
			rq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_0) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.1 Safari/605.1.15")
			//rq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
			//rsp, err := httputil.RQUntil(http.DefaultClient, rq)
			rsp, err := http.DefaultClient.Do(rq)
			if err != nil {
				//proxyErr(src, fmt.Errorf("rq for page with list: %s", err))
				continue
			}
			page, err := goquery.NewDocumentFromReader(rsp.Body)
			if err != nil {
				//proxyErr(src, fmt.Errorf("parse page HTML: %s", err))
				continue
			}
			rsp.Body.Close()

			rawList := strings.Split(page.Find("div.modal-body > textarea").Get(0).FirstChild.Data, "\n\n")[1]
			for _, line := range strings.Split(strings.TrimSpace(rawList), "\n") {
				Add(proxy.New(line))
			}

			<-t.C
		}
	}

	addFuncs[src] = run
}
