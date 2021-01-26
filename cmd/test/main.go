package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"wallmask/internal/platform/db"
	"wallmask/pkg/proxy"
)

func main() {
	rs := db.Query(`
				SELECT id, protocol, ipv4, port
				FROM proxies
				WHERE working = $1
				ORDER BY lastTested;`, false)
	fmt.Println(rs.Next())
}

func maind() {
	cli := &http.Client{
		Transport: &http.Transport{
			Proxy:           proxy.Rand(),
			TLSClientConfig: &tls.Config{},
		},
	}

	switch {
	}
	for {
		rsp, err := cli.Get("https://bytesimal.github.io/test")
		if err != nil {
			log.Print(false)
			continue
		}
		bd, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Print(false)
			continue
		}
		log.Print(bytes.Contains(bd, []byte("TEST PAGE")))
		time.Sleep(500 * time.Millisecond)
	}
}
