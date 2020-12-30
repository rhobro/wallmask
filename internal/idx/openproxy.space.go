package idx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Bytesimal/goutils/pkg/coll"
	"github.com/Bytesimal/goutils/pkg/fileio"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"wallmask/pkg/proxy"
)

func init() {
	src := "openproxy.space"
	run := func() {
		base := "https://api.openproxy.space/list?"
		listBase := "https://api.openproxy.space/list/"
		refreshDuration := 12 * time.Hour
		t := time.NewTicker(refreshDuration)

		// structs for json unmarshal
		type index struct {
			Protocols []int  `json:"protocols"`
			Count     int    `json:"count"`
			Code      string `json:"code"`
			Time      int64  `json:"date"`
		}
		type countryList struct {
			Anons     []int `json:"anons"`
			Countries []struct {
				URLs []string `json:"items"`
			} `json:"data"`
		}
		type plainList struct {
			Anons []int    `json:"anons"`
			Data  []string `json:"data"`
		}
		// planners
		var latest int64
		var firstIndexed bool

		// Index all proxies
		v := url.Values{}
		v.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1000000, 10))

		// Loop to get index urls and keep looping for first page after 1 full idx run through
		var n int
		for {
			// Get params for index urls
			v.Set("skip", strconv.Itoa(n))

			if firstIndexed {
				<-t.C
			}

			// Request
			rq, _ := http.NewRequest("GET", base+v.Encode(), nil)
			rq.Header.Set("User-Agent", httputil.RandUA())
			rsp, err := httputil.RQUntil(http.DefaultClient, rq)
			if err != nil {
				proxyErr(src, fmt.Errorf("rq for lists: %s", err))
				continue
			}
			bd, err := ioutil.ReadAll(rsp.Body)
			if err != nil {
				proxyErr(src, fmt.Errorf("reading rq body for lists: %s", err))
				continue
			}
			rsp.Body.Close()

			// Parse JSON
			var lists []index
			err = json.Unmarshal(bd, &lists)
			if err != nil {
				// Save json in tmp file for post debugging
				f, fErr := ioutil.TempFile(fileio.TmpDir, "openproxy.space_json_*.json")

				if fErr != nil {
					proxyErr(src, fmt.Errorf("creating temp file at %s: %s\n", fileio.TmpDir, err))
				} else {
					_, cErr := io.Copy(f, bytes.NewReader(bd))

					if cErr != nil {
						proxyErr(src, fmt.Errorf("copying json into tmpfile at %s: %s\n", f.Name(), cErr))
					} else {
						proxyErr(src, fmt.Errorf("unmarshaling proxy lists, json at %s: %s", f.Name(), err))
					}
					f.Close()
				}

				continue
			}

			// stop loop if all have been indexed
			if len(lists) == 0 {
				n = 0
				firstIndexed = true
				continue
			}

			for _, lSrc := range lists {
				// Update latest
				if lSrc.Time > latest {
					latest = lSrc.Time
				} else {
					if firstIndexed {
						continue
					}
				}

				// Choose http(s) proxies
				if coll.ContainsInt(lSrc.Protocols, 1) || coll.ContainsInt(lSrc.Protocols, 2) &&
					lSrc.Count > 0 {
					// get JSON for countryList
					rq, _ := http.NewRequest("GET", listBase+lSrc.Code, nil)
					rq.Header.Set("User-Agent", httputil.RandUA())
					rsp, err := httputil.RQUntil(http.DefaultClient, rq)
					if err != nil {
						proxyErr(src, fmt.Errorf("rq proxies in countryList: %s", err))
						continue
					}
					bd, err := ioutil.ReadAll(rsp.Body)
					if err != nil {
						proxyErr(src, fmt.Errorf("reading proxies in countryList: %s", err))
						continue
					}
					rsp.Body.Close()

					// Parse JSON
					var l countryList
					err = json.Unmarshal(bd, &l)
					if err != nil {
						// Try use plainList
						var l plainList
						err = json.Unmarshal(bd, &l)
						if err != nil {
							// If both formats don't work
							// Save json in tmp file for post debugging
							f, fErr := ioutil.TempFile(fileio.TmpDir, "openproxy.space_json_*.json")

							if fErr != nil {
								proxyErr(src, fmt.Errorf("creating temp file at %s: %s\n", fileio.TmpDir, err))
							} else {
								_, cErr := io.Copy(f, bytes.NewReader(bd))

								if cErr != nil {
									proxyErr(src, fmt.Errorf("copying json into tmpfile at %s: %s\n", f.Name(), cErr))
								} else {
									proxyErr(src, fmt.Errorf("unmarshaling proxies from countryList, json at %s: %s", f.Name(), err))
								}
								f.Close()
							}

							continue
						}

						// Index from plain list
						for _, raw := range l.Data {
							Add(proxy.New(raw))
						}
					}

					// Go through lists
					for _, country := range l.Countries {
						for _, raw := range country.URLs {
							Add(proxy.New(raw))
						}
					}
				}
			}

			if !firstIndexed {
				n += len(lists)
			}
		}
	}

	addFuncs[src] = run
}
