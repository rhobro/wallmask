package idx

import (
	"bufio"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/Bytesimal/goutils/pkg/util"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func init() {
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
			rsp, err := http.DefaultClient.Do(rq)
			util.Check(err)
			rd := bufio.NewReader(rsp.Body)

			for {
				line, err := rd.ReadString('\n')
				line = strings.TrimSpace(line)
				// Check for EOF
				if err != nil {
					break
				}

				// add after parsing string
				Add(New(line))
			}

			<-t.C
		}
	}

	src := "proxyscrape.com"
	addFuncs[src] = run
}
