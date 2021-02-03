package consts

import (
	configcat "github.com/configcat/go-sdk/v7"
	"time"
)

const (
	ConfigCatKey = "W8fYCA2kvkeP3BUhP-sxcg/-KMXy47nX0KhQsrTB1_sCg"
)

var ConfigCatConf = configcat.Config{
	SDKKey:       ConfigCatKey,
	PollingMode:  configcat.AutoPoll,
	PollInterval: 1 * time.Second,
}
