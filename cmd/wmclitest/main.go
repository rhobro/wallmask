package main

import (
	"bytes"
	"crypto/tls"
	"github.com/rhobro/wallmask/pkg/proxy"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	proxy.Init(25)
	var cli *http.Client

	for {
		cli = &http.Client{
			Transport: &http.Transport{
				Proxy:           proxy.Rand(),
				TLSClientConfig: &tls.Config{},
			},
		}

		s := time.Now()
		rsp, err := cli.Get("https://whatismyipaddress.com/proxy-check")
		e := time.Now()
		if err != nil {
			log.Print(err)
			continue
		}
		bd, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			log.Print(err)
			continue
		}
		log.Printf("Succeeded: %t | Time: %s", bytes.Contains(bd, []byte("TEST PAGE")), e.Sub(s))
	}
}
