package idx

import (
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"time"
	"wallmask/pkg/core"
)

func init() {
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
				continue
			}
			page, err := goquery.NewDocumentFromReader(rsp.Body)
			if err != nil {
				continue
			}

			// Process
			raw := strings.TrimSpace(page.Find("textarea#rawData").Get(0).FirstChild.Data)
			for _, line := range strings.Split(raw, "\n") {
				Add(core.New(line))
			}

			<-t.C
		}
	}

	src := "coderduck.com"
	addFuncs[src] = run
}
