package main

import (
	"bufio"
	configcat "github.com/configcat/go-sdk/v7"
	"github.com/getsentry/sentry-go"
	"github.com/rhobro/goutils/pkg/fileio"
	"github.com/rhobro/goutils/pkg/services/cfgcat"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/internal/idx"
	"github.com/rhobro/wallmask/internal/platform/consts"
	"github.com/rhobro/wallmask/internal/platform/db"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func init() {
	// tmp files
	fileio.Init("", "wmidx")
	// cfgcat
	cfgcat.InitCustom(configcat.Config{
		SDKKey: consts.ConfigCatKey,
		Transport: &http.Transport{
			MaxIdleConns: 1,
		},
	}, true)
	// sentry
	sentree.Init(sentry.ClientOptions{
		Dsn:              cfgcat.C.GetStringValue("sentryDSN", "", nil),
		AttachStacktrace: true,
		Environment:      "server",
	}, true)
	// db
	db.Connect()
}

func main() {
	// Start indexing
	idx.Index()

	// Wait until user clicks enter
	rd := bufio.NewScanner(os.Stdin)
	for rd.Scan() {
		if strings.Contains(rd.Text(), "q") {
			break
		} else {
			log.Printf("n Goroutines: %d", runtime.NumGoroutine())
		}
	}

	// close
	cfgcat.C.Close()
	db.Close()
}
