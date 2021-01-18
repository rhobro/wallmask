package main

import (
	"fmt"
	"github.com/Bytesimal/goutils/pkg/httputil"
	"time"
)

func main() {
	const nSamples = 100

	var durs []time.Duration
	for i := 0; i < nSamples; i++ {
		s := time.Now()
		httputil.RandUA()
		durs = append(durs, time.Since(s))
	}
	fmt.Println(durs)
	var tot time.Duration
	for _, d := range durs {
		tot += d
	}
	fmt.Println(tot / nSamples)
}
