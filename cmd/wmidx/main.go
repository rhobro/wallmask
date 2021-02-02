package main

import (
	"bufio"
	configcat "github.com/configcat/go-sdk/v7"
	"github.com/rhobro/goutils/pkg/cfgcat"
	"github.com/rhobro/goutils/pkg/fileio"
	"github.com/rhobro/wallmask/internal/idx"
	"github.com/rhobro/wallmask/internal/platform/consts"
	"github.com/rhobro/wallmask/internal/platform/db"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

func init() {
	// tmp files
	fileio.Init("", "wmidx")
	// cfgcat
	cfgcat.InitCustom(configcat.Config{
		SDKKey:       consts.ConfigCatKey,
		PollingMode:  configcat.AutoPoll,
		PollInterval: 1 * time.Second,
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
}
