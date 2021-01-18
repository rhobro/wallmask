package idx

import (
	"bytes"
	"crypto/tls"
	"github.com/jackc/pgx/v4"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"wallmask/internal/platform/db"
	"wallmask/pkg/proxy"
)

type idx struct {
	Period time.Duration
	Last   time.Time
	F      func()
}

var idxrs = make(map[string]*idx)

func Index() {
	// launch idx scheduler
	go scheduler()
	// launch db testers
	//go dbTest(true)
	//go dbTest(false)

	log.Println("{proxy} Initialized")
}

func scheduler() {
	for src, i := range idxrs {
		log.Printf("{proxy} indexing %s", src)
		i.F()
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
		WHERE ipv4 = $1 AND port = $2
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
			db.Exec(sqlUpdate, d.ID, d.Last, d.Ok)
		}
	}
}

func AddBatch(ps []*proxy.Proxy) {
	ds := batchDetails(ps)
	b := pgx.Batch{}
	for i, d := range ds {
		if d.ID == -1 {
			// Add to database if not in already
			b.Queue(sqlInsert, ps[i].Protocol, ps[i].IPv4, ps[i].Port, d.Last, d.Ok)
		} else {
			// Update last tested if already in db
			b.Queue(sqlUpdate, d.Last, d.Ok, d.ID)
		}
	}
	db.BatchExec(&b)
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
	rs := db.Query(sqlDetails, p.IPv4, p.Port)
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

func batchDetails(ps []*proxy.Proxy) (ds []*detailStruct) {
	b := pgx.Batch{}
	ds = make([]*detailStruct, len(ps), len(ps))

	for _, p := range ps {
		b.Queue(sqlDetails, p.IPv4, p.Port)
	}
	rss := db.BatchQuery(&b)

	// read batch
	for i, rs := range rss {
		// Get id if present
		var id int64 = -1
		if rs.Next() {
			err := rs.Scan(&id)
			if err != nil {
				log.Printf("scan result set of count occurences of %s: %s", ps[i], err)
			}
		}
		ds[i] = &detailStruct{
			ID: id,
		}
	}

	// batch test
	for i, p := range ps {
		last, ok := test(p)
		ds[i].Last = last
		ds[i].Ok = ok
	}
	return
}

const dbTestF = 10 * time.Second

func dbTest(working bool) {
	t := time.NewTicker(dbTestF)
	for {
		// test Proxies
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
				continue
			}

			last, ok := test(&p)
			db.Exec(sqlUpdate, id, last, ok)
		}
		rs.Close()

		<-t.C
	}
}

const (
	testTimeout        = 1 * time.Second
	maxConcurrentTests = 250
)

var testCli = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{},
	},
	Timeout: testTimeout,
}
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

var testSemaphore = make(chan struct{}, maxConcurrentTests)

func testRQ(p *proxy.Proxy) (lastTested time.Time, ok bool) {
	u, err := p.URL()
	if err == nil {
		testCli.Transport.(*http.Transport).Proxy = http.ProxyURL(u)

		testSemaphore <- struct{}{}
		rsp, err := testCli.Get("https://bytesimal.github.io/test")
		<-testSemaphore
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
