package idx

import (
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"net/http"
	"time"
)

func init() {
	run := func() {
		base := "https://www.my-proxy.com/free-elite-proxy.html"
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

			// Recursively extract elements
			igMyProxyExtract(page.Find("div.list").Get(0).FirstChild)
			igMyProxyExtract(page.Find("div.to-lock").Get(0).FirstChild)

			<-t.C
		}
	}

	src := "my-proxy.com"
	addFuncs[src] = run
}

func igMyProxyExtract(n *html.Node) {
	// Process text
	if n.Data != "br" {
		Add(New(n.Data))
	}

	// Move onto next sibling
	if n.NextSibling != nil && n.NextSibling.Data != "div" {
		igMyProxyExtract(n.NextSibling)
	}
}
