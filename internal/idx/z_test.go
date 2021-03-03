package idx

import (
	"github.com/rhobro/wallmask/internal/platform"
	"github.com/rhobro/wallmask/pkg/proxy"
	"sync"
	"testing"
)

func TestIndexers(t *testing.T) {
	platform.Init()

	// proxy list
	var pMu sync.Mutex
	proxies := make(map[string]*proxy.Proxy)
	// subtest test dir
	var sMu sync.RWMutex
	subtests := make(map[string]*testing.T)

	// modified in-memory add
	Add = func(p *proxy.Proxy) {
		pMu.Lock()
		if p != nil {
			proxies[p.String()] = p
		}
		pMu.Unlock()
	}
	// modified proxy err to report to test
	proxyErr = func(src string, err error) {
		sMu.RLock()
		subtests[src].Error(err)
		sMu.RUnlock()
	}

	// test
	for src, i := range idxrs {
		src := src // capture loop variables
		i := i
		t.Run(src, func(t *testing.T) {
			// prep
			sMu.Lock()
			subtests[src] = t
			sMu.Unlock()
			t.Parallel()

			// run func
			i.F()
		})
	}
}
