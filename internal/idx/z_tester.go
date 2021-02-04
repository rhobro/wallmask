package idx

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/internal/platform/db"
	"github.com/rhobro/wallmask/pkg/proxy"
	"io/ioutil"
	"net/http"
	"time"
)

const (
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

// to control number of test workers started
const nTestWorkers = 100

// for deciding if the select statements should order ASC or DESC
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

func dbTest(working bool, order sqlOrder, limit int) {
	for {
		// test Proxies
		var rs pgx.Rows
		if limit == -1 {
			rs = db.Query(fmt.Sprintf(`
				SELECT id, protocol, ipv4, port
				FROM proxies
				WHERE working = $1
				ORDER BY lastTested %s;`, order), working)
		} else {
			rs = db.Query(fmt.Sprintf(`
				SELECT id, protocol, ipv4, port
				FROM proxies
				WHERE working = $1
				ORDER BY lastTested %s
				LIMIT %d;`, order, limit), working)
		}

		// get proxy and test
		for rs.Next() {
			var p proxy.Proxy
			var id int64
			err := rs.Scan(&id, &p.Protocol, &p.IPv4, &p.Port)
			if err != nil {
				sentree.LogCaptureErr(fmt.Errorf("querying proxies for update test: %s", err))
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

// To reduce impact of anomalous proxy test fail
const nTestRetries = 5
const retryInterval = 1 * time.Second

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
			time.Sleep(retryInterval)
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
				MaxIdleConns:    1,
			},
			Timeout: testTimeout,
		}
		defer cli.CloseIdleConnections()

		rsp, err := cli.Get("https://rhobro.github.io/test")
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
		sentree.LogCaptureErr(fmt.Errorf("{proxy} can't parse url of proxy %s: %s", p.String(), err))
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
			sentree.LogCaptureErr(fmt.Errorf("can't test parse html: %s", err))
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
				sentree.LogCaptureErr(fmt.Errorf("can't parse bool %s: %s", rawBool, err))
			}
			positive = positive && testResult
		})

		ok = visibleIP != pubIP // && !positive TODO to only allow anonymous proxiesz

	} else {
		sentree.LogCaptureErr(fmt.Errorf("{proxy} can't parse url of proxy %s: %s", p.String(), err))
	}
	return
}*/
