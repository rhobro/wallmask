package platform

import (
	"github.com/getsentry/sentry-go"
	"github.com/rhobro/goutils/pkg/fileio"
	"github.com/rhobro/goutils/pkg/services/cfgcat"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/internal/platform/consts"
	"github.com/rhobro/wallmask/internal/platform/db"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func Init() {
	cInit("server")
}

func InitCli() {
	cInit("client")
}

func InitTest() {
	// cfgcat
	cfgcat.InitCustom(consts.ConfigCatConfig, true)
	// sentry
	sentree.Init(sentry.ClientOptions{
		Dsn:              cfgcat.C.GetStringValue("sentryDSN", "", nil),
		AttachStacktrace: true,
		Environment:      "test",
		HTTPTransport: &http.Transport{
			MaxIdleConns: 1,
		},
	}, true)
	// tmp files
	fileio.Init("", "wmidx")

	rand.Seed(time.Now().UnixNano())
}

func cInit(env string) {
	// cfgcat
	cfgcat.InitCustom(consts.ConfigCatConfig, true)
	// sentry
	sentree.Init(sentry.ClientOptions{
		Dsn:              cfgcat.C.GetStringValue("sentryDSN", "", nil),
		AttachStacktrace: true,
		Environment:      env,
		HTTPTransport: &http.Transport{
			MaxIdleConns: 1,
		},
	}, true)
	// db
	db.Connect(true)
	// tmp files
	fileio.Init("", "wmidx")

	if env != "test" {
		rand.Seed(time.Now().UnixNano())
	}
}

func Close() {
	fileio.Close()
	db.Close()
	cfgcat.C.Close()
	os.Exit(0)
}
