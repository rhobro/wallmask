package consts

import (
	"github.com/configcat/go-sdk/v7"
	"net/http"
	"time"
)

const (
	ConfigCatKey = "W8fYCA2kvkeP3BUhP-sxcg/-KMXy47nX0KhQsrTB1_sCg"
)

var ConfigCatConfig = configcat.Config{
	SDKKey: ConfigCatKey,
	Transport: &http.Transport{
		MaxIdleConns: 1,
	},
	PollInterval: 1 * time.Second,
}
