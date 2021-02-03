package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/rhobro/goutils/pkg/httputil"
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
	cfgcat.InitCustom(consts.ConfigCatConf, true)
	// sentry
	sentree.Init(sentry.ClientOptions{
		Dsn:              cfgcat.C.GetStringValue("sentryDSN", "", nil),
		AttachStacktrace: true,
		Environment:      "Client Tester",
	}, true)
	proxy.Init(25)
}

func main() {
	rq, _ := http.NewRequest("GET", "https://rhobro.github.io/test", nil)

	for {
		http.DefaultClient = &http.Client{
			Transport: &http.Transport{
				Proxy:           proxy.Rand(),
				TLSClientConfig: &tls.Config{},
			},
		}

		s := time.Now()
		rsp, err := httputil.RQUntilCustom(http.DefaultClient, rq, -1)
		e := time.Now()
		if err != nil {
			sentree.LogCaptureErr(fmt.Errorf("requesting page: %s", err))
			continue
		}
		bd, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			sentree.LogCaptureErr(fmt.Errorf("reading page bytes: %s", err))
			continue
		}
		log.Printf("Succeeded: %t | Time: %s", bytes.Contains(bd, []byte("TEST PAGE")), e.Sub(s))
	}
}
