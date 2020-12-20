package idx

import (
	"encoding/json"
	"github.com/Bytesimal/goutils/pkg/coll"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"wallmask/pkg/core"
)

func init() {
	run := func() {
		base := "https://api.openproxy.space/list?"
		listBase := "https://api.openproxy.space/list/"
		refreshDuration := 12 * time.Hour
		t := time.NewTicker(refreshDuration)

		// structs for json unmarshalling
		type listSrc struct {
			Protocols []int  `json:"protocols"`
			Count     int    `json:"count"`
			Code      string `json:"code"`
			Time      int64  `json:"date"`
		}
		// planners
		var latest int64
		var firstIdxd bool

		// Index all proxies
		v := url.Values{}
		v.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1000000, 10))

		// Loop to get listSrc urls and keep looping for first page after 1 full idx run through
		var n int
		for {
			// Get params for listSrc urls
			v.Set("skip", strconv.Itoa(n))

			if firstIdxd {
				<-t.C
			}

			// Request
			rq, _ := http.NewRequest("GET", base+v.Encode(), nil)
			rq.Header.Set("User-Agent", httputil.RandUA())
			rsp, err := httputil.RQUntil(http.DefaultClient, rq)
			if err != nil {
				continue
			}
			bd, err := ioutil.ReadAll(rsp.Body)
			if err != nil {
				continue
			}
			rsp.Body.Close()

			// Parse JSON
			var lists []listSrc
			err = json.Unmarshal(bd, &lists)
			if err != nil {
				continue
			}

			// stop loop if all have been indexed
			if len(lists) == 0 {
				n = 0
				firstIdxd = true
				continue
			}

			for _, lSrc := range lists {
				// Update latest
				if lSrc.Time > latest {
					latest = lSrc.Time
				} else {
					if firstIdxd {
						continue
					}
				}

				// Choose http(s) proxies
				if coll.ContainsInt(lSrc.Protocols, 1) || coll.ContainsInt(lSrc.Protocols, 2) &&
					lSrc.Count > 0 {
					// get JSON for list
					rq, _ := http.NewRequest("GET", listBase+lSrc.Code, nil)
					rq.Header.Set("User-Agent", httputil.RandUA())
					rsp, err := httputil.RQUntil(http.DefaultClient, rq)
					if err != nil {
						continue
					}
					bd, err := ioutil.ReadAll(rsp.Body)
					if err != nil {
						continue
					}
					rsp.Body.Close()

					// Parse JSON
					type list struct {
						Anons     []int `json:"anons"`
						Countries []struct {
							URLs []string `json:"items"`
						} `json:"data"`
					}
					var l list
					err = json.Unmarshal(bd, &l)
					if err != nil {
						continue
					}

					// Go through lists
					for _, country := range l.Countries {
						for _, raw := range country.URLs {
							Add(core.New(raw))
						}
					}
				}
			}

			if !firstIdxd {
				n += len(lists)
			}
		}
	}

	src := "openproxy.space"
	addFuncs[src] = run
}
