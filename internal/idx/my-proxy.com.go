package idx

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/wallmask/pkg/proxy"
	"golang.org/x/net/html"
	"net/http"
	"strings"
	"time"
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
			proxyErr(src, err)
			return
		}
		defer rsp.Body.Close()
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		if err != nil {
			proxyErr(src, err)
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

		p, err := proxy.New(d)
		if err == nil {
			Add(p)
		}
	}

	// Move onto next sibling
	if n.NextSibling != nil && n.NextSibling.Data != "div" {
		recursiveExtract(proto, n.NextSibling)
	}
}
