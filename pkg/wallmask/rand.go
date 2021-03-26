package wallmask

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/rhobro/goutils/pkg/services/cfgcat"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/internal/platform/consts"
	"github.com/rhobro/wallmask/internal/platform/db"
	"log"
	"net/http"
	"net/url"
	"sync"
)

var bufSize = 25
var verb bool
var proxies chan *url.URL

func Rand() func(*http.Request) (*url.URL, error) {
	return func(r *http.Request) (u *url.URL, _ error) {
		initOnce.Do(initialize)
		return <-proxies, nil
	}
}

func Init(bSize int, verbose bool) {
	// cfgcat
	cfgcat.InitCustom(consts.ConfigCatConfig, verbose)
	// sentry
	sentree.Init(sentry.ClientOptions{
		Dsn:              cfgcat.C.GetStringValue("sentryDSN", "", nil),
		AttachStacktrace: true,
		Environment:      "client",
		HTTPTransport: &http.Transport{
			MaxIdleConns: 1,
		},
	}, verbose)
	// db
	db.Connect(verbose)

	bufSize = bSize
	verb = verbose
	initOnce.Do(initialize)
}

// for lazy initialization
var initOnce sync.Once

func initialize() {
	proxies = make(chan *url.URL, bufSize)
	go func() {
		for {
			// repeatedly query
			rs := db.Query(fmt.Sprintf(`
				SELECT protocol || '://' || ipv4 || ':' || CAST(port AS TEXT) ip
				FROM proxies
				WHERE working AND protocol != '' AND ipv4 != ''
				ORDER BY lastTested DESC
				LIMIT %d;`, bufSize))

			// loop through each proxy
			for rs.Next() {
				var p string
				err := rs.Scan(&p)
				if err != nil {
					sentree.C.CaptureException(err, nil, nil)
					log.Printf("can't loop through proxy table rows: %s", err)
				}
				u, err := url.Parse(p)
				if err == nil && u != nil {
					proxies <- u
				} else {
					sentree.C.CaptureException(err, nil, nil)
					log.Printf("can't parse url %s: %s", p, err)
				}
			}
		}
	}()

	// wait till until 1 entry
	for len(proxies) == 0 {
		continue
	}
	if verb {
		log.Print("{wallmask} connected")
	}
}
