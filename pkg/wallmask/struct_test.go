package wallmask

import (
	"fmt"
	"math"
	"math/rand"
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
		switch rand.Intn(4) {
		// abnormal proxies - random letters
		case 0:
			var text string
			for i := 0; i < rand.Intn(20); i++ {
				text += chars[rand.Intn(len(chars))]
			}
			testSet[text] = nil

		// normal proxies
		default:
			ip := fmt.Sprintf("%d.%d.%d.%d",
				rand.Intn(math.MaxUint8),
				rand.Intn(math.MaxUint8),
				rand.Intn(math.MaxUint8),
				rand.Intn(math.MaxUint8))
			port := rand.Intn(math.MaxUint16-1) + 1

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
