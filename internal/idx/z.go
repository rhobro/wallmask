package idx

import (
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"wallmask/internal/platform/db"
	"wallmask/pkg/proxy"
)

var addFuncs = make(map[string]func())

func Index() {
	// launch indexers
	for src, f := range addFuncs {
		log.Printf("{proxy} indexing %s", src)
		go f()
	}

	// launch db testers
	//go dbTest(true)
	//go dbTest(false)

	log.Println("{proxy} Initialized")
}

func Add(p *proxy.Proxy) {
	if p != nil {
		// Add if positive test with working bool
		last, ok := test(p)

		// check if already exists
		rs := db.Query(`
			SELECT id
			FROM proxies
			WHERE ipv4 = $1 AND port = $2
			LIMIT 1;`, p.IPv4, p.Port)
		defer rs.Close()

		// Get id if present
		var id int64 = -1
		if rs.Next() {
			err := rs.Scan(&id)
			if err != nil {
				log.Printf("scan result set of count occurences of %s: %s", p, err)
			}
		}

		if id == -1 {
			// Add to database if not in already
			db.Exec(`
				INSERT INTO proxies (ipv4, port, lastTested, working)
				VALUES ($1, $2, $3, $4);`, p.IPv4, p.Port, last, ok)
		} else {
			// Update last tested if already in db
			update(id, last, ok)
		}
	}
}

const (
	testTimeout        = 1 * time.Second
	maxConcurrentTests = 250
)

var semaphore = make(chan struct{}, maxConcurrentTests)

func test(p *proxy.Proxy) (lastTested time.Time, ok bool) {
	u, err := p.URL()
	if err == nil {
		cli := http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(u),
				TLSClientConfig: &tls.Config{},
			},
			Timeout: testTimeout,
		}

		// test request
		semaphore <- struct{}{}
		rsp, err := cli.Get("https://bytesimal.github.io/test")
		<-semaphore
		lastTested = time.Now()

		if err == nil {
			defer rsp.Body.Close()
			bd, err := ioutil.ReadAll(rsp.Body)
			if err == nil {
				// Check response contains test text
				if strings.Contains(string(bd), "TEST PAGE") {
					ok = true
				}
			}
		}
	}
	return
}

func update(id int64, last time.Time, working bool) {
	// Update lastChecked and working
	db.Exec(`
		UPDATE proxies
		SET lastTested = $1, working = $2
		WHERE id = $3;`, last, working, id)
}

const dbTestF = 1 * time.Minute

func dbTest(working bool) {
	t := time.NewTicker(dbTestF)
	for {
		// test proxies
		rs := db.Query(`
				SELECT id, ipv4, port
				FROM proxies
				WHERE working = $1
				ORDER BY lastTested;`, working)

		// get proxy and test
		for rs.Next() {
			var p proxy.Proxy
			var id int64
			err := rs.Scan(&id, &p.IPv4, &p.Port)
			if err != nil {
				log.Printf("querying proxies for update test: %s", err)
			}

			last, ok := test(&p)
			update(id, last, ok)
		}
		rs.Close()

		<-t.C
	}
}

func proxyErr(src string, err error) {
	log.Printf("{proxy} {%s} %s", src, err)
}
