package platform

import (
	"github.com/getsentry/sentry-go"
	"github.com/rhobro/goutils/pkg/fileio"
	"github.com/rhobro/goutils/pkg/services/cfgcat"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/internal/platform/consts"
	"math/rand"
	"time"
)

func Init() {
	cInit(false)
}

func InitTest() {
	cInit(true)
}

func cInit(test bool) {
	var env string
	if test {
		env = "test"
	} else {
		env = "server"
	}

	// tmp files
	fileio.Init("", "wmidx")
	// cfgcat
	cfgcat.InitCustom(consts.ConfigCatConfig, true)
	// sentry
	sentree.Init(sentry.ClientOptions{
		Dsn:              cfgcat.C.GetStringValue("sentryDSN", "", nil),
		AttachStacktrace: true,
		Environment:      env,
	}, true)

	if !test {
		rand.Seed(time.Now().UnixNano())
	}
}
