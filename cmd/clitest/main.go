package main

import (
	"bytes"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/pkg/wallmask"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func init() {
	wallmask.Init(25, true)
}

func main() {
	rq, _ := http.NewRequest("GET", "https://rhobro.github.io/test", nil)

	for {
		http.DefaultClient = &http.Client{
			Transport: &http.Transport{
				Proxy: wallmask.Rand(),
			},
		}

		s := time.Now()
		rsp, err := http.DefaultClient.Do(rq)
		e := time.Now()
		if err != nil {
			sentree.C.CaptureException(err, nil, nil)
			log.Printf("requesting page: %s", err)
			continue
		}
		bd, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			sentree.C.CaptureException(err, nil, nil)
			log.Printf("reading page bytes: %s", err)
			continue
		}
		log.Printf("Succeeded: %t | Time: %s", bytes.Contains(bd, []byte("TEST PAGE")), e.Sub(s))
	}
}
