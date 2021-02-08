package main

import (
	"bufio"
	"github.com/getsentry/sentry-go"
	"github.com/rhobro/goutils/pkg/fileio"
	"github.com/rhobro/goutils/pkg/services/cfgcat"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/internal/idx"
	"github.com/rhobro/wallmask/internal/platform/consts"
	"github.com/rhobro/wallmask/internal/platform/db"
	"log"
	"os"
	"runtime"
	"strings"
)

func init() {
	// tmp files
	fileio.Init("", "wmidx")
	// cfgcat
	cfgcat.InitCustom(consts.ConfigCatConfig, true)
	// sentry
	sentree.Init(sentry.ClientOptions{
		Dsn:              cfgcat.C.GetStringValue("sentryDSN", "", nil),
		AttachStacktrace: true,
		Environment:      "server",
	}, true)
	// db
	db.Connect(true)
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
