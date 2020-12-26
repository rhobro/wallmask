package idx

import (
	"crypto/tls"
	"fmt"
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
	/*go func() {
		for {
			dbProxiesTest(true)
		}
	}()
	go func() {
		for {
			dbProxiesTest(false)
		}
	}()*/
	log.Println("{proxy} Initialized")
}

func Add(p *proxy.Proxy) {
	if p != nil {
		// Add if positive test with working bool
		last, ok := test(p)

		if ok {
			fmt.Println(p)
		}
		return

		// check if already exists
		rs, err := db.Exec(fmt.Sprintf(`
					SELECT id
					FROM prx_proxies
					WHERE ipv4 = '%s' AND port = %d;`, p.IPv4, p.Port))
		if err != nil {
			log.Printf("count occurences of %s: %s", p, err)
			return
		}

		// Get id if present
		var id int
		if rs.Next() {
			err := rs.Scan(&id)
			if err != nil {
				log.Printf("scan result set of count occurences of %s: %s", p, err)
			}
		}
		rs.Close()

		if id == 0 {
			// Add to database if not
			_, err = db.Exec(fmt.Sprintf(`
						INSERT INTO prx_proxies (ipv4, port, lastTested, working)
						VALUES ('%s', %d, %d, %t);`, p.IPv4, p.Port, last, ok))
			if err != nil {
				log.Printf("add %s to db: %s", p, err)
			}
		} else {
			// Update last tested if already in
			update(id, last, ok)
		}
	}
}

const testTimeout = 1 * time.Second
const maxConcurrentTests = 100

var semaphore = make(chan struct{}, maxConcurrentTests)

func test(p *proxy.Proxy) (lastTested int64, ok bool) {
	return 0, true
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
		lastTested = time.Now().Unix()

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
	// test proxies which are currently working
	rs, err := db.Exec(fmt.Sprintf(`
		SELECT id, ipv4, port
		FROM prx_proxies
		WHERE working = %t
		ORDER BY lastTested;`, working))
	if err != nil {
		log.Printf("listing proxy from db to update test: %s", err)
	}

	// test and update
	for rs.Next() {
		var p proxy.Proxy
		var id int
		err := rs.Scan(&id, &p.IPv4, &p.Port)
		if err != nil {
			log.Printf("querying proxies for update test: %s", err)
		}

		last, ok := test(&p)
		update(id, last, ok)
	}
	rs.Close()
}

func update(id int, last int64, working bool) {
	// Update lastChecked
	_, err := db.Exec(fmt.Sprintf(`
						UPDATE prx_proxies
						SET lastTested = %d, working = %t
						WHERE id = %d;`, last, working, id))
	if err != nil {
		log.Printf("update proxy in db with id %d: %s", id, err)
	}
}

func proxyErr(src string, err error) {
	log.Printf("{proxy} {%s} %s", src, err)
}
