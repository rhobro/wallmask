package platform

import (
	"github.com/getsentry/sentry-go"
	"github.com/rhobro/goutils/pkg/fileio"
	"github.com/rhobro/goutils/pkg/services/cfgcat"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/internal/platform/consts"
)

func Init() {
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
}

func InitTest() {
	// tmp files
	fileio.Init("", "wmidx")
	// cfgcat
	cfgcat.InitCustom(consts.ConfigCatConfig, true)
	// sentry
	sentree.Init(sentry.ClientOptions{
		Dsn:              cfgcat.C.GetStringValue("sentryDSN", "", nil),
		AttachStacktrace: true,
		Environment:      "test",
	}, true)
}
