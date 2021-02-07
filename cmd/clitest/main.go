package main

import (
	"bytes"
	"github.com/getsentry/sentry-go"
	"github.com/rhobro/goutils/pkg/services/cfgcat"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/internal/platform/consts"
	"github.com/rhobro/wallmask/pkg/proxy"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func init() {
	// cfgcat
	cfgcat.InitCustom(consts.ConfigCatConfig, true)
	// sentry
	sentree.Init(sentry.ClientOptions{
		Dsn:              cfgcat.C.GetStringValue("sentryDSN", "", nil),
		AttachStacktrace: true,
		Environment:      "clitester",
	}, true)
	proxy.Init(25)
}

func main() {
	rq, _ := http.NewRequest("GET", "https://rhobro.github.io/test", nil)

	for {
		http.DefaultClient = &http.Client{
			Transport: &http.Transport{
				Proxy: proxy.Rand(),
			},
		}

		s := time.Now()
		//rsp, err := httputil.RQUntilCustom(http.DefaultClient, rq, -1)
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
