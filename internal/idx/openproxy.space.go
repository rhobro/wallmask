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
	base, _ := url.Parse("https://api.openproxy.space/list")
	listBase := "https://api.openproxy.space/list/"
	// Planners
	var latest int64
	var firstIdxd bool

	// structs
	type index struct {
		Protocols     []int  `json:"protocols"`
		Anons         []int  `json:"anons"`
		Code          string `json:"code"`
		WithCountries bool   `json:"withCountries"`
		Date          int64  `json:"date"`
	}
	type countryList struct {
		Anons     []int `json:"anons"`
		Countries []struct {
			Proxies []string `json:"items"`
		} `json:"data"`
	}
	type longList struct {
		Anons   []int    `json:"anons"`
		Proxies []string `json:"data"`
	}

	// constants
	const (
		HTTP = iota + 1
		HTTPS
		SOCKS5 = iota + 2
	)

	run := func() {
		var skip int
		for {
			base.RawQuery = rawQuery(skip)
			rq, _ := http.NewRequest("GET", base.String(), nil)
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

			// Parse JSON lists of lists
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

			if len(lists) == 0 {
				firstIdxd = true
				break
			}
			skip += len(lists)

			// iterate through index
			for _, l := range lists {
				var sch proxy.Protocol
				if coll.ContainsInt(l.Protocols, HTTP) || coll.ContainsInt(l.Protocols, HTTPS) {
					sch = proxy.HTTP
				} else if coll.ContainsInt(l.Protocols, SOCKS5) {
					sch = proxy.SOCKS5
				}

				if sch != "" {
					// update planners
					if l.Date > latest {
						latest = l.Date
					} else {
						if firstIdxd {
							return
						}
					}

					rq, _ := http.NewRequest("GET", listBase+l.Code, nil)
					rq.Header.Set("User-Agent", httputil.RandUA())
					rsp, err := httputil.RQUntil(http.DefaultClient, rq)
					if err != nil {
						proxyErr(src, fmt.Errorf("rq proxies in list %s: %s", l.Code, err))
						continue
					}
					bd, err := ioutil.ReadAll(rsp.Body)
					if err != nil {
						proxyErr(src, fmt.Errorf("reading proxies in list: %s", err))
						continue
					}

					// Unmarshal
					if l.WithCountries {
						var cList countryList
						err := json.Unmarshal(bd, &cList)
						if err != nil {
							proxyErr(src, fmt.Errorf("unmarshaling with countries list: %s", err))
						}

						// add
						for _, country := range cList.Countries {
							for _, raw := range country.Proxies {
								p := proxy.New(raw)
								p.Protocol = sch
								Add(p)
							}
						}

					} else {
						var list longList
						err := json.Unmarshal(bd, &list)
						if err != nil {
							proxyErr(src, fmt.Errorf("unmarshaling without countries list: %s", err))
						}

						// add
						for _, raw := range list.Proxies {
							p := proxy.New(raw)
							p.Protocol = sch
							Add(p)
						}
					}
				}
			}
		}
	}

	idxrs[src] = &idx{
		Period: 12 * time.Hour,
		run:    run,
	}
}

func rawQuery(skip int) string {
	v := url.Values{}
	v.Set("skip", strconv.Itoa(skip))
	v.Set("ts", strconv.FormatInt(time.Now().UnixNano()/1000000, 10))
	return v.Encode()
}
