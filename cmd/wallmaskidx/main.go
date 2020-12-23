package main

import (
	_ "github.com/ibmdb/go_ibm_db"
	"net/http"
	"net/url"
	"wallmask/pkg/proxy"
)

func init() {
	dbg, _ := url.Parse("http://localhost:9090")
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(dbg),
		},
	}
}

func main() {
	f := proxy.Rand()
	f(nil)
}
