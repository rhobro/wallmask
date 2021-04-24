package idx

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/goutils/pkg/services/cfgcat"
	"github.com/rhobro/wallmask/pkg/wallmask"
	"io/ioutil"
	"net/http"
	"time"
)

func init() {
	src := "shiftytr.github"

	bases := map[string]wallmask.Protocol{
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/%s/http.txt":   wallmask.HTTP,
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/%s/https.txt":  wallmask.HTTPS,
		"https://raw.githubusercontent.com/ShiftyTR/Proxy-List/%s/socks5.txt": wallmask.SOCKS5,
	}
	// planners
	var firstIdxd bool
	var latestSHA string

	scrape := func(sha string) {
		for file, proto := range bases {
			// Request
			rq, _ := http.NewRequest("GET", fmt.Sprintf(file, sha), nil)
			rq.Header.Set("User-Agent", httputil.RandUA())
			rsp, err := httputil.RQUntil(http.DefaultClient, rq)
			if err != nil {
				proxyErr(src, err)
				return
			}
			rd := bufio.NewScanner(rsp.Body)

			for rd.Scan() {
				text, err := rd.Text(), rd.Err()
				if err != nil {
					continue
				}
				p, err := wallmask.New(text)
				if err != nil {
					continue
				}
				p.Proto = proto
				Add(p)
			}
		}
	}

	run := func(isTest bool) {
		branchesAPI := fmt.Sprintf("https://%s@api.github.com/repos/ShiftyTR/Proxy-List/branches", cfgcat.C.GetStringValue("ghOAuthToken", "", nil))
		commitsAPI := fmt.Sprintf("https://%s@api.github.com/repos/ShiftyTR/Proxy-List/commits?per_page=100", cfgcat.C.GetStringValue("ghOAuthToken", "", nil))
		commitsAPI += "&sha=%s"

		// list branchesAPI for latest commit
		rq, _ := http.NewRequest("GET", branchesAPI, nil)
		rq.Header.Set("User-Agent", httputil.RandUA())
		rsp, err := httputil.RQUntil(http.DefaultClient, rq)
		if err != nil {
			proxyErr(src, err)
			return
		}
		defer rsp.Body.Close()
		// unmarshal
		bd, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			proxyErr(src, err)
			return
		}
		var branches []struct {
			Name   string `json:"name"`
			Commit struct {
				SHA string `json:"sha"`
			} `json:"commit"`
		}
		err = json.Unmarshal(bd, &branches)
		if err != nil {
			proxyErr(src, err)
			return
		}

		if firstIdxd {
			// loop through branches until "master" found and get latest sha
			for _, b := range branches {
				if b.Name == "master" {
					if b.Commit.SHA != latestSHA {
						scrape(b.Commit.SHA)
						latestSHA = b.Commit.SHA
					}
					break
				}
			}

		} else {
			// loop through branches until "master" found and get latest sha
			for _, b := range branches {
				if b.Name == "master" {
					latestSHA = b.Commit.SHA
					break
				}
			}

			// for each sha, scrape
			lastSHA := latestSHA
			for {
				// list commits since last
				rq, _ := http.NewRequest("GET", fmt.Sprintf(commitsAPI, lastSHA), nil)
				rq.Header.Set("User-Agent", httputil.RandUA())
				rsp, err := httputil.RQUntil(http.DefaultClient, rq)
				if err != nil {
					proxyErr(src, err)
					continue
				}
				bd, err := ioutil.ReadAll(rsp.Body)
				if err != nil {
					proxyErr(src, err)
					continue
				}
				rsp.Body.Close()

				var commits []struct{ SHA string }
				err = json.Unmarshal(bd, &commits)
				if err != nil {
					proxyErr(src, err)
					continue
				}
				if len(commits) == 1 {
					break
				}

				// loop through and index SHAs
				for i, s := range commits[1:] {
					scrape(s.SHA)
					if i == len(commits)-2 {
						lastSHA = s.SHA
					}

					if isTest {
						return
					}
				}
			}

			firstIdxd = true
		}
	}

	idxrs[src] = &idx{
		Period: 15 * time.Minute,
		run:    run,
	}
}
