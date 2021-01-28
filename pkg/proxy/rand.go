package proxy

import (
	"log"
	"net/http"
	"net/url"
	"sync"
	"wallmask/internal/platform/db"
)

var bufSize = 25

func Init(bSize int) {
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
					log.Printf("can't loop through proxy rows: %s", err)
				}
				u, err := url.Parse(p)
				if err == nil {
					proxies <- u
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
	return func(r *http.Request) (u *url.URL, err error) {
		initOnce.Do(initialize)
		return <-proxies, nil
	}
}
