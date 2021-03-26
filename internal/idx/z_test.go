package idx

import (
	"github.com/rhobro/goutils/pkg/services/sentree"
	"github.com/rhobro/wallmask/internal/platform"
	"github.com/rhobro/wallmask/pkg/wallmask"
	"sync"
	"testing"
)

func TestIndexers(t *testing.T) {
	platform.InitTest()

	// proxy list
	var pMu sync.Mutex
	proxies := make(map[string]*wallmask.Proxy)
	// subtest test dir
	var sMu sync.RWMutex
	subtests := make(map[string]*testing.T)

	// modified in-memory add
	Add = func(p *wallmask.Proxy) {
		pMu.Lock()
		if p != nil {
			proxies[p.String()] = p
		}
		pMu.Unlock()
	}
	// modified proxy err to report to test
	proxyErr = func(src string, err error) {
		// subtest error
		sMu.RLock()
		subtests[src].Error(err)
		sMu.RUnlock()

		// capture via sentry
		sentree.C.CaptureException(err, nil, nil)
	}

	// test
	for src, i := range idxrs {
		src, i := src, i // capture loop variables

		// run parallel test
		t.Run(src, func(t *testing.T) {
			// prep
			sMu.Lock()
			subtests[src] = t
			sMu.Unlock()
			t.Parallel()

			// run func
			i.F(true)
		})
	}
}
