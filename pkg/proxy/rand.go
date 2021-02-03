package proxy

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

func Init(bSize int) {
	// cfgcat
	cfgcat.InitCustom(consts.ConfigCatConf, false)
	// sentry
	sentree.Init(sentry.ClientOptions{
		Dsn:              cfgcat.C.GetStringValue("sentryDSN", "", nil),
		AttachStacktrace: true,
		Environment:      "Client",
	}, true)
	// db
	db.Connect()

	bufSize = bSize
	initOnce.Do(initialize)
}

// for lazy initialization
var initOnce sync.Once

func initialize() {
	proxies = make(chan *url.URL, bufSize)
	go func() {
		for {
			// repeatedly query
			rs := db.Query(`
				SELECT protocol || '://' || ipv4 || ':' || CAST(port AS TEXT) ip
				FROM proxies
				WHERE working AND protocol != '' AND ipv4 != ''
				ORDER BY lastTested DESC;`)

			// loop through each proxy
			for rs.Next() {
				var p string
				err := rs.Scan(&p)
				if err != nil {
					sentree.LogCaptureErr(fmt.Errorf("can't loop through proxy table rows: %s", err))
				}
				u, err := url.Parse(p)
				if err == nil {
					proxies <- u
				} else {
					sentree.LogCaptureErr(fmt.Errorf("can't parse url %s: %s", p, err))
				}
			}
		}
	}()

	// wait till until 1 entry
	for len(proxies) == 0 {
		continue
	}
	log.Print("{wallmask} connected")
}

var proxies chan *url.URL

func Rand() func(*http.Request) (*url.URL, error) {
	return func(r *http.Request) (*url.URL, error) {
		initOnce.Do(initialize)
		return <-proxies, nil
	}
}
