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
	go func() {
		for {
			dbProxiesTest(true)
		}
	}()
	go func() {
		for {
			dbProxiesTest(false)
		}
	}()
	log.Println("{proxy} Initialized")
}

func Add(p *proxy.Proxy) {
	if p != nil {
		// Add if positive test with working bool
		last, ok := test(p)

		// check if already exists
		rs, err := db.Query(`
			SELECT id
			FROM proxies
			WHERE ipv4 = $1 AND port = $2;`, p.IPv4, p.Port)
		if err != nil {
			log.Printf("count occurences of %s: %s", p, err)
			return
		}
		defer rs.Close()

		// Get id if present
		var id int64
		if rs.Next() {
			err := rs.Scan(&id)
			if err != nil {
				log.Printf("scan result set of count occurences of %s: %s", p, err)
			}
		}

		if id == 0 {
			// Add to database if not
			err := db.Exec(`
				INSERT INTO proxies (ipv4, port, lastTested, working)
				VALUES ($1, $2, $3, $4);`, p.IPv4, p.Port, last, ok)
			if err != nil {
				log.Printf("add %s to db: %s", p, err)
			}
		} else {
			// Update last tested if already in db
			update(id, last, ok)
		}
	}
}

const testTimeout = 1 * time.Second
const maxConcurrentTests = 100

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

func dbProxiesTest(working bool) {
	// test proxies
	rs, err := db.Query(`
		SELECT id, ipv4, port
		FROM proxies
		WHERE working = $1
		ORDER BY lastTested;`, working)
	if err != nil {
		log.Printf("listing proxy from db to update test: %s", err)
	}

	// test and update
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
}

func update(id int64, last time.Time, working bool) {
	// Update lastChecked
	err := db.Exec(`
		UPDATE proxies
		SET lastTested = $1, working = $2
		WHERE id = $3;`, last, working, id)
	if err != nil {
		log.Printf("update proxy in db with id %d: %s", id, err)
	}
}

func proxyErr(src string, err error) {
	log.Printf("{proxy} {%s} %s", src, err)
}
