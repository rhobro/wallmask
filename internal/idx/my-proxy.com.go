package idx

import (
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "my-proxy.com"
	bases := map[proxy.Protocol]string{
		proxy.HTTP:   "https://www.my-proxy.com/free-elite-proxy.html",
		proxy.SOCKS5: "https://www.my-proxy.com/free-socks-5-proxy.html",
	}

	scrape := func(proto proxy.Protocol, base string) {
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
			proxyErr(src, fmt.Errorf("parse page with goquery: %s", err))
			return
		}

		// Recursively extract elements
		recursiveExtract(proto, page.Find("div.list").Get(0).FirstChild)
		recursiveExtract(proto, page.Find("div.to-lock").Get(0).FirstChild)
	}

	run := func() {
		for sc, ur := range bases {
			scrape(sc, ur)
		}
	}

	idxrs[src] = &idx{
		Period: 10 * time.Minute,
		run:    run,
	}
}

func recursiveExtract(proto proxy.Protocol, n *html.Node) {
	// Process text
	if n.Data != "br" {
		d := n.Data
		hashI := strings.Index(d, "#")
		if hashI > -1 {
			d = d[:hashI]
		}

		p := proxy.New(d)
		if p != nil {
			p.Protocol = proto
			Add(p)
		}
	}

	// Move onto next sibling
	if n.NextSibling != nil && n.NextSibling.Data != "div" {
		recursiveExtract(proto, n.NextSibling)
	}
}
