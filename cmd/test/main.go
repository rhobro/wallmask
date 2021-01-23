package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wallmask/pkg/proxy"
)

func main() {
	base := "https://www.proxynova.com/proxy-server-list/elite-proxies/"
	run := func() {
		// Request
		rq, _ := http.NewRequest("GET", base, nil)
		rq.Header.Set("User-Agent", httputil.RandUA())
		rsp, err := httputil.RQUntil(http.DefaultClient, rq)
		if err != nil {
			log.Printf("rq for list page: %s", err)
			return
		}
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		if err != nil {
			log.Printf("parse page HTML: %s", err)
			return
		}
		rsp.Body.Close()

		page.Find("table.table > tbody > tr").Each(func(i int, sl *goquery.Selection) {
			if sl.Find("td > abbr > script").Length() > 0 {
				ip := sl.Find("td > abbr > script").Get(0).FirstChild.Data
				ip = ip[strings.Index(ip, "'")+1:]
				ip = ip[:strings.Index(ip, "'")]

				port, err := strconv.Atoi(strings.TrimSpace(sl.Find("td").Get(1).FirstChild.Data))
				if err != nil {
					return
				}

				fmt.Println(&proxy.Proxy{
					IPv4: ip,
					Port: uint16(port),
				})
			}
		})
	}
	run()
}

func maind() {
	cli := &http.Client{
		Transport: &http.Transport{
			Proxy:           proxy.Rand(),
			TLSClientConfig: &tls.Config{},
		},
	}

	switch {
	}
	for {
		rsp, err := cli.Get("https://bytesimal.github.io/test")
		if err != nil {
			log.Print(false)
			continue
		}
		bd, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Print(false)
			continue
		}
		log.Print(bytes.Contains(bd, []byte("TEST PAGE")))
		time.Sleep(500 * time.Millisecond)
	}
}
