package main

import (
	"bufio"
	"github.com/Bytesimal/goutils/pkg/fileio"
	"log"
	"os"
	"runtime"
	"strings"
	"wallmask/internal/idx"
)

func init() {
	fileio.Init("", "wmidx")
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
