package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
	"wallmask/internal/idx"
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
	for {
		fmt.Println(len(idx.Proxies))
		time.Sleep(time.Second)
	}
}
