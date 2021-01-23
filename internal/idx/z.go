package idx

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"wallmask/internal/platform/db"
	"wallmask/pkg/proxy"
)

type idx struct {
	Period time.Duration
	run    func()

	last    time.Time
	running bool
}

// presumes scheduler has set running and last
func (i *idx) F() {
	i.run()
	i.running = false
}

var idxrs = make(map[string]*idx)

const nTestWorkers = 10

func Index() {
	// launch idx scheduler
	go scheduler()
	log.Println("{proxy} Initialized")
}

func scheduler() {
	// launch db testers
	go dbTest(true)
	go dbTest(false)
	for i := 0; i < nTestWorkers; i++ {
		go dbTestWorker()
	}

	for {
		for _, i := range idxrs {
			if time.Since(i.last) > i.Period && !i.running {
				i.last = time.Now()
				i.running = true
				go i.F()
			}
		}
	}
}

const (
	sqlInsert = `
		INSERT INTO proxies (protocol, ipv4, port, lastTested, working)
		VALUES ($1, $2, $3, $4, $5);`
	sqlUpdate = `
		UPDATE proxies
		SET lastTested = $1, working = $2
		WHERE id = $3;`
	sqlDetails = `
		SELECT id
		FROM proxies
		WHERE protocol = $1 AND ipv4 = $2 AND port = $3
		LIMIT 1;`
)

func Add(p *proxy.Proxy) {
	if p != nil {
		d := details(p)
		if d.ID == -1 {
			// Add to database if not in already
			db.Exec(sqlInsert, p.Protocol, p.IPv4, p.Port, d.Last, d.Ok)
		} else {
			// Update last tested if already in db
			db.Exec(sqlUpdate, d.Last, d.Ok, d.ID)
		}
	}
}

type detailStruct struct {
	ID   int64
	Last time.Time
	Ok   bool
}

func details(p *proxy.Proxy) *detailStruct {
	// test
	last, ok := test(p)

	// check if already exists
	rs := db.Query(sqlDetails, p.Protocol, p.IPv4, p.Port)
	defer rs.Close()

	// Get id if present
	var id int64 = -1
	if rs.Next() {
		err := rs.Scan(&id)
		if err != nil {
			log.Printf("scan result set of count occurences of %s: %s", p, err)
		}
	}
	return &detailStruct{
		ID:   id,
		Last: last,
		Ok:   ok,
	}
}

type testInst struct {
	ID int64
	P  *proxy.Proxy
}

var testPipe = make(chan testInst, maxConcurrentTests)

func dbTest(working bool) {
	for {
		// test Proxies
		rs := db.Query(`
				SELECT id, protocol, ipv4, port
				FROM proxies
				WHERE working = $1
				ORDER BY lastTested;`, working)

		// get proxy and test
		for rs.Next() {
			var p proxy.Proxy
			var id int64
			err := rs.Scan(&id, &p.Protocol, &p.IPv4, &p.Port)
			if err != nil {
				log.Printf("querying proxies for update test: %s", err)
				continue
			}

			testPipe <- testInst{
				ID: id,
				P:  &p,
			}
		}
		rs.Close()
	}
}

func dbTestWorker() {
	for i := range testPipe {
		last, ok := test(i.P)
		db.Exec(sqlUpdate, last, ok, i.ID)
	}
}

const (
	testTimeout        = 1 * time.Second
	maxConcurrentTests = 250
)

var protocols = []proxy.Protocol{proxy.HTTP, proxy.SOCKS5}

func test(p *proxy.Proxy) (lastTested time.Time, ok bool) {
	// if no protocol
	if p.Protocol == "" {
		// test each protocol
		for _, sc := range protocols {
			p.Protocol = sc
			// test
			lastTested, ok = testRQ(p)
			if ok {
				break
			}
		}
		// nil proto if none - must test later
		if !ok {
			p.Protocol = ""
		}

	} else {
		lastTested, ok = testRQ(p)
	}

	return
}

//var testSemaphore = make(chan struct{}, maxConcurrentTests)

func testRQ(p *proxy.Proxy) (lastTested time.Time, ok bool) {
	u, err := p.URL()
	if err == nil {
		cli := &http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(u),
				TLSClientConfig: &tls.Config{},
			},
			Timeout: testTimeout,
		}

		//testSemaphore <- struct{}{}
		rsp, err := cli.Get("https://bytesimal.github.io/test")
		//<-testSemaphore
		lastTested = time.Now()
		if err != nil {
			return
		}
		defer rsp.Body.Close()
		bd, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Printf("can't read rsp of test page with proxy %s: %s", p, err)
			return
		}
		ok = bytes.Contains(bd, []byte("TEST PAGE"))

	} else {
		log.Printf("{proxy} can't parse url of proxy %s: %s", p.String(), err)
	}
	return
}

func proxyErr(src string, err error) {
	log.Printf("{proxy} {%s} %s", src, err)
}
