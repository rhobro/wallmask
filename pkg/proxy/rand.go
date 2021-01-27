package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"wallmask/internal/platform/db"
)

const defBufSize = 25

func Init(bufSize int) {
	if bufSize < 0 {
		panic(fmt.Sprintf("buffer size for proxies is too small: %d", bufSize))
	}

	proxyStream = make(chan string, bufSize)
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
				proxyStream <- p
			}
		}
	}()

	// wait till until 1 entry
	for len(proxyStream) == 0 {
		continue
	}
	log.Print("{wallmask} connected")
}

var proxyStream chan string

func Rand() func(*http.Request) (*url.URL, error) {
	return func(r *http.Request) (u *url.URL, err error) {
		if proxyStream == nil {
			Init(defBufSize)
		}
		s := <-proxyStream
		return url.Parse(s)
	}
}
