package main

import (
	"net/http"
	"net/url"
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

}
