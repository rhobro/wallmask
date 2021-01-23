package idx

import (
	"fmt"
	"github.com/Bytesimal/goutils/pkg/fileio"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "coderduck.com"
	base := "https://www.coderduck.com/free-proxy-list"

	run := func() {
		// Request
		rq, _ := http.NewRequest("GET", base, nil)
		rq.Header.Set("User-Agent", httputil.RandUA())
		rsp, err := httputil.RQUntil(http.DefaultClient, rq)
		if err != nil {
			proxyErr(src, fmt.Errorf("rq for page list page: %s", err))
			return
		}
		defer rsp.Body.Close()
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		if err != nil {
			proxyErr(src, fmt.Errorf("parse HTML from page: %s", err))
			return
		}

		// Process
		sl := page.Find("textarea#rawData")
		if sl.Length() > 0 {
			raw := strings.TrimSpace(sl.Get(0).FirstChild.Data)
			for _, line := range strings.Split(raw, "\n") {
				Add(proxy.New(line))
			}
		} else {
			f, err := ioutil.TempFile(fileio.TmpDir, "coderduck.com_*.html")
			if err != nil {
				log.Printf("can't create tmp file: %s", err)
				return
			}
			defer f.Close()
			pgHTML, err := page.Html()
			if err != nil {
				log.Printf("can't export doc to HTML str: %s", err)
			}
			_, err = f.WriteString(pgHTML)
			if err != nil {
				log.Printf("can't write html to tmp file: %s", err)
			}
		}
	}

	//idxrs[src] = &idx{
	_ = &idx{
		Period: 10 * time.Minute,
		run:    run,
	}
}
