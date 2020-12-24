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
	src := "coderduck.com"
	run := func() {
		base := "https://www.coderduck.com/free-proxy-list"
		refreshDuration := 10 * time.Minute
		t := time.NewTicker(refreshDuration)

		for {
			// Request
			rq, _ := http.NewRequest("GET", base, nil)
			rq.Header.Set("User-Agent", httputil.RandUA())
			rsp, err := httputil.RQUntil(http.DefaultClient, rq)
			if err != nil {
				proxyErr(src, fmt.Errorf("rq for page with list: %s", err))
				continue
			}
			page, err := goquery.NewDocumentFromReader(rsp.Body)
			if err != nil {
				proxyErr(src, fmt.Errorf("parse HTML from page: %s", err))
				continue
			}

			// Process
			raw := strings.TrimSpace(page.Find("textarea#rawData").Get(0).FirstChild.Data)
			for _, line := range strings.Split(raw, "\n") {
				Add(proxy.New(line))
			}

			<-t.C
		}
	}

	addFuncs[src] = run
}
