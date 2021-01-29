package idx

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"wallmask/internal/platform/db"
	"wallmask/pkg/proxy"
)

var idxrs = make(map[string]*idx)

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

func Index() {
	// launch idx scheduler
	go scheduler()
	log.Print("{proxy} Initialized")
}

func scheduler() {
	// launch db testers
	go dbTest(true, ASC)
	go dbTest(false, DESC)
	for i := 0; i < nTestWorkers; i++ {
		go testWorker()
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
		if httputil.IsValidIPv4(p.IPv4) {
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
}

type detail struct {
	ID   int64
	Last time.Time
	Ok   bool
}

func details(p *proxy.Proxy) *detail {
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
	return &detail{
		ID:   id,
		Last: last,
		Ok:   ok,
	}
}

// To ensure that proxies are not over-tested repeatedly
const (
	nTestRetries = 10
	nTestWorkers = 50
)

type sqlOrder string

const (
	ASC  sqlOrder = "ASC"
	DESC sqlOrder = "DESC"
)

type testInst struct {
	ID      int64
	P       *proxy.Proxy
	Working bool
}

var testPipe = make(chan *testInst, nTestWorkers)

func dbTest(working bool, order sqlOrder) {
	for {
		// test Proxies
		rs := db.Query(fmt.Sprintf(`
				SELECT id, protocol, ipv4, port
				FROM proxies
				WHERE working = $1
				ORDER BY lastTested %s;`, order), working)

		// get proxy and test
		for rs.Next() {
			var p proxy.Proxy
			var id int64
			err := rs.Scan(&id, &p.Protocol, &p.IPv4, &p.Port)
			if err != nil {
				log.Printf("querying proxies for update test: %s", err)
				continue
			}

			testPipe <- &testInst{
				ID:      id,
				P:       &p,
				Working: working,
			}
		}
		rs.Close()
	}
}

func testWorker() {
	for ti := range testPipe {
		// test with optional retries
		var last time.Time
		var ok bool
		for i := 0; i < nTestRetries; i++ {
			last, ok = test(ti.P)
			if ok {
				break
			}
		}

		if ok || ti.Working {
			db.Exec(sqlUpdate, last, ok, ti.ID) // only update if positive test or if working proxy fails
		}
	}
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
				return
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

const testTimeout = 1 * time.Second

//var pubIP string

// Get public ip to check with proxy tests
/*func init() {
	rsp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		log.Fatalf("can't get public ip: %s", err)
	}
	defer rsp.Body.Close()
	bd, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Fatalf("can't get public ip: %s", err)
	}

	type ipRsp struct {
		IP string `json:"ip"`
	}
	var ip ipRsp
	err = json.Unmarshal(bd, &ip)
	if err != nil {
		log.Fatalf("can't get public ip: %s", err)
	}
	pubIP = ip.IP
}*/

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
		defer cli.CloseIdleConnections()

		rsp, err := cli.Get("https://bytesimal.github.io/test")
		lastTested = time.Now()
		if err != nil {
			return
		}
		defer rsp.Body.Close()
		bd, err := ioutil.ReadAll(rsp.Body)
		if err != nil {

			return
		}
		ok = bytes.Contains(bd, []byte("TEST PAGE"))

	} else {
		log.Printf("{proxy} can't parse url of proxy %s: %s", p.String(), err)
	}
	return
}

/*func testRQ(p *proxy.Proxy) (lastTested time.Time, ok bool) { TODO test with anonymous check
	u, err := p.URL()
	if err == nil {
		cli := &http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(u),
				TLSClientConfig: &tls.Config{},
			},
			Timeout: testTimeout,
		}
		defer cli.CloseIdleConnections()

		rsp, err := cli.Get("https://whatismyipaddress.com/proxy-check")
		lastTested = time.Now()
		if err != nil {
			return
		}
		defer rsp.Body.Close()
		page, err := goquery.NewDocumentFromReader(rsp.Body)
		if err != nil {
			log.Printf("can't test parse html: %s", err)
			return
		}

		var visibleIP string
		var positive bool
		page.Find("table > tbody > tr").Each(func(i int, sl *goquery.Selection) {
			if i == 0 {
				// get displayed ip
				visibleIP = sl.Find("td").Get(1).FirstChild.Data
				return
			}

			// get bools from table
			rawBool := sl.Find("td > span").Text()
			testResult, err := strconv.ParseBool(rawBool)
			if err != nil {
				log.Printf("can't parse bool %s: %s", rawBool, err)
			}
			positive = positive && testResult
		})

		ok = visibleIP != pubIP // && !positive TODO to only allow anonymous proxiesz

	} else {
		log.Printf("{proxy} can't parse url of proxy %s: %s", p.String(), err)
	}
	return
}*/

func proxyErr(src string, err error) {
	log.Printf("{proxy} {%s} %s", src, err)
}
