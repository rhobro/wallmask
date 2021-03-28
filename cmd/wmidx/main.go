package main

import (
	"bufio"
	"github.com/rhobro/goutils/pkg/services/cfgcat"
	"github.com/rhobro/wallmask/internal/idx"
	"github.com/rhobro/wallmask/internal/platform"
	"github.com/rhobro/wallmask/internal/platform/db"
	"log"
	"os"
	"runtime"
	"strings"
)

func init() {
	platform.Init()
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
