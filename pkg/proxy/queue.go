package proxy

import (
	"log"
	"net/http"
	"net/url"
	"wallmask/internal/platform/db"
	"wallmask/pkg/core"
)

const qSize = 25

var q = make(chan *core.Proxy, qSize)

func init() {
	// Test if db already init
	st, err := db.DB.Prepare(`
		SELECT ipv4, port
		FROM prx_proxies
		ORDER BY lastTested DESC;`)
	if err != nil {
		log.Fatalf("can't access prx_proxies table: %s\n", err)
	}
	defer st.Close()

	rs, err := db.Exec(st)
	if err != nil {
		log.Fatalf("can't access prx_proxies table exec: %s\n", err)
	}
	defer rs.Close()
}

func Rand() func(r *http.Request) (*url.URL, error) {
	return func(*http.Request) (*url.URL, error) {
		return (<-q).URL()
	}
}
