package idx

import (
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"net/http"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "my-proxy.com"
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
				proxyErr(src, fmt.Errorf("rq for page with list: %s", err))
				continue
			}
			page, err := goquery.NewDocumentFromReader(rsp.Body)
			if err != nil {
				proxyErr(src, fmt.Errorf("parse page with goquery: %s", err))
				continue
			}

			// Recursively extract elements
			igMyProxyExtract(page.Find("div.list").Get(0).FirstChild)
			igMyProxyExtract(page.Find("div.to-lock").Get(0).FirstChild)

			<-t.C
		}
	}

	addFuncs[src] = run
}

func igMyProxyExtract(n *html.Node) {
	// Process text
	if n.Data != "br" {
		Add(proxy.New(n.Data))
	}

	// Move onto next sibling
	if n.NextSibling != nil && n.NextSibling.Data != "div" {
		igMyProxyExtract(n.NextSibling)
	}
}
