package idx

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/rhobro/goutils/pkg/fileio"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/pkg/proxy"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func init() {
	src := "coderduck.com"
	base := "https://www.coderduck.com/free-proxy-list"

	run := func(bool) {
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

		// Process
		sl := page.Find("textarea#rawData")
		if sl.Length() > 0 {
			raw := strings.TrimSpace(sl.Get(0).FirstChild.Data)
			for _, line := range strings.Split(raw, "\n") {
				p, err := proxy.New(line)
				if err == nil {
					Add(p)
				}
			}
		} else {
			f, err := ioutil.TempFile(fileio.TmpDir, "coderduck.com_*.html")
			if err != nil {
				sentree.C.CaptureException(err, nil, nil)
				log.Printf("can't create tmp file: %s", err)
				return
			}
			defer f.Close()
			pgHTML, err := page.Html()
			if err != nil {
				sentree.C.CaptureException(err, nil, nil)
				log.Printf("can't export doc to HTML str: %s", err)
			}
			_, err = f.WriteString(pgHTML)
			if err != nil {
				sentree.C.CaptureException(err, nil, nil)
				log.Printf("can't write html to tmp file: %s", err)
			}
		}
	}

	idxrs[src] = &idx{
		Period: 10 * time.Minute,
		run:    run,
	}
}
