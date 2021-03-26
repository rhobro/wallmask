package wallmask

import (
	"fmt"
	"github.com/rhobro/goutils/pkg/util"
	"math"
	"strings"
	"testing"
)

// TODO disabled
func benchmarkNewProxy(t *testing.B) {
	// setup
	chars := strings.Split("abcdefghijklmnopqrstuvwxyz1234567890.,:/", "")

	// prepare test examples
	testSet := make(map[string]*Proxy)
	for i := 0; i < t.N; i++ {
		switch util.Rand.Intn(4) {
		// abnormal proxies - random letters
		case 0:
			var text string
			for i := 0; i < util.Rand.Intn(20); i++ {
				text += chars[util.Rand.Intn(len(chars))]
			}
			testSet[text] = nil

		// normal proxies
		default:
			ip := fmt.Sprintf("%d.%d.%d.%d",
				util.Rand.Intn(math.MaxUint8),
				util.Rand.Intn(math.MaxUint8),
				util.Rand.Intn(math.MaxUint8),
				util.Rand.Intn(math.MaxUint8))
			port := util.Rand.Intn(math.MaxUint16-1) + 1

			testSet[fmt.Sprintf("%s:%d", ip, port)] = &Proxy{
				IPv4: ip,
				Port: uint16(port),
			}
		}
	}

	t.ResetTimer()
	// test
	for in, expected := range testSet {
		out, err := New(in)

		// expect nil
		if expected == nil {
			if out != expected {
				t.Errorf("failed on in: %s | expected out: %s | out: %s | err: %s", in, expected, out, err)
			}
		} else {
			if err != nil {
				t.Errorf("failed on in: %s | expected out: %s | out: %s | err: %s", in, expected, out, err)
				continue
			}

			if *out != *expected {
				t.Errorf("failed on in: %s | expected out: %s | out: %s | err: %s", in, expected, out, err)
			}
		}
	}
}
