package idx

import (
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/rhobro/goutils/pkg/httputil"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/internal/platform/db"
	"github.com/rhobro/wallmask/pkg/proxy"
	"log"
	"time"
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

func scheduler() {
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

const sqlInsert = `
	INSERT INTO proxies (protocol, ipv4, port, lastTested, working)
	VALUES ($1, $2, $3, $4, $5);`

// add func made into a variable for testing purposes
var Add = func(p *proxy.Proxy) {
	if p != nil {
		if httputil.IsValidIPv4(p.IPv4) {
			d := details(p)
			if d.ID == -1 {
				// Add to database if not in already
				db.Exec(sqlInsert, p.Proto, p.IPv4, p.Port, d.Last, d.Ok)
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
	rs := db.QueryRow(sqlDetails, p.Proto, p.IPv4, p.Port)

	// Get id if present
	var id int64 = -1
	err := rs.Scan(&id)
	if err != nil && err != pgx.ErrNoRows {
		sentree.C.CaptureException(err, nil, nil)
		log.Print(fmt.Errorf("scan result set of count occurences of %s: %s", p, err))
	}

	return &detail{
		ID:   id,
		Last: last,
		Ok:   ok,
	}
}

var proxyErr = func(src string, err error) {
	sentree.C.CaptureException(err, nil, nil)
	log.Printf("{proxy} {%s} %s", src, err)
}
