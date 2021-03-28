package main

import (
	"bufio"
	"github.com/rhobro/wallmask/internal/idx"
	"github.com/rhobro/wallmask/internal/platform"
	"log"
	"os"
	"runtime"
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
		if rd.Text() == "q" {
			break

		} else {
			log.Printf("n Goroutines: %d", runtime.NumGoroutine())
		}
	}

	// close
	platform.Close()
}
