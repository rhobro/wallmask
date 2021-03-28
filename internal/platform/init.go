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
	"time"
)

func Init() {
	cInit("server")
}

func InitCli() {
	cInit("client")
}

func InitTest() {
	cInit("test")
}

func cInit(env string) {
	// tmp files
	fileio.Init("", "wmidx")
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

	if env != "test" {
		rand.Seed(time.Now().UnixNano())
	}
}
