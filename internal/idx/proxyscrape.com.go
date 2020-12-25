package idx

import (
	"bufio"
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "proxyscrape.com"
	run := func() {
		base := "https://api.proxyscrape.com/v2/?"
		refreshDuration := 5 * time.Minute
		t := time.NewTicker(refreshDuration)

		for {
			// Params
			v := url.Values{}
			v.Set("request", "displayproxies")
			v.Set("protocol", "http")
			v.Set("timeout", "10000") // TODO change to 50
			v.Set("country", "all")
			v.Set("ssl", "all")
			v.Set("anonymity", "elite")
			v.Set("simplified", "false")

			// Request
			rq, _ := http.NewRequest("GET", base+v.Encode(), nil)
			rq.Header.Set("User-Agent", httputil.RandUA())
			rsp, err := httputil.RQUntil(http.DefaultClient, rq)
			if err != nil {
				proxyErr(src, fmt.Errorf("rq proxy list: %s", err))
				continue
			}
			rd := bufio.NewReader(rsp.Body)

			for {
				line, err := rd.ReadString('\n')
				// Check for EOF or error
				if err != nil {
					if err != io.EOF {
						proxyErr(src, fmt.Errorf("reading list: %s", err))
					}
					break
				}
				line = strings.TrimSpace(line)

				// add after parsing string
				Add(proxy.New(line))
			}
			rsp.Body.Close()

			<-t.C
		}
	}

	addFuncs[src] = run
}
