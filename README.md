# wallmask

[![Build Status](https://travis-ci.com/rhobro/wallmask.svg?token=uVJBazFqkUpqe5ptFH7x&branch=master)](https://travis-ci.com/rhobro/wallmask)

[![codecov](https://codecov.io/gh/rhobro/wallmask/branch/master/graph/badge.svg?token=CC751QEJAV)](https://codecov.io/gh/rhobro/wallmask)

[![CodeScene Code Health](https://codescene.io/projects/12745/status-badges/code-health)](https://codescene.io/projects/12745)
[![CodeScene System Mastery](https://codescene.io/projects/12745/status-badges/system-mastery)](https://codescene.io/projects/12745)
[![CodeScene general](https://codescene.io/images/analyzed-by-codescene-badge.svg)](https://codescene.io/projects/12745)

[![BCH compliance](https://bettercodehub.com/edge/badge/rhobro/wallmask?branch=master&token=d50754a1e026a62374a2ffa97d70d31c2d2c168d)](https://bettercodehub.com/)

[![DeepSource](https://deepsource.io/gh/rhobro/wallmask.svg/?label=active+issues&show_trend=true&token=kBOTwUVbUDau-jma24Sd51ad)](https://deepsource.io/gh/rhobro/wallmask/?ref=repository-badge)
[![DeepSource](https://deepsource.io/gh/rhobro/wallmask.svg/?label=resolved+issues&show_trend=true&token=kBOTwUVbUDau-jma24Sd51ad)](https://deepsource.io/gh/rhobro/wallmask/?ref=repository-badge)

After realising the extent of tracking using cookies and IPs on the internet, I became more privacy conscious and was
looking for a way to anonymize my web activity. Existing solutions such as Tor lack speed due to many layers of proxies,
while standalone proxies were insecure since all requests could be traced to that proxy IP. My solution aims to mitigate
the drawbacks of both the aforementioned solutions.

### Design

It has two components, the server and client side.

#### Server

The "indexer" consists of multiple web scrapers which are scheduled based on when they were last run and the frequency
at which they should run. These run in parallel to scrape multiple sites with proxy lists. These proxies are then tested
with a test page hosted on GitHub Pages at https://rhobro.github.io/test. The proxy is then added to an SQL database (
currently using PostgreSQL with CockroachDB), where it registers whether it was working when tested and the timestamp of
the test.

The "tester" consists of multiple worker goroutines which listen on a channel of proxies. Three goroutines
asynchronously test the dead proxies, recently died proxies and working proxies. These add their goroutines into the
channel to be tested and updated by the workers.

#### Client

This consists of a goroutine which repeatedly queries the database to return the latest set of working proxies which it
adds to a channel. Whenever the proxy function is called, it returns the latest proxy url in the channel, blocking until
another one is added to the channel.

### Usage

The client package is very lightweight and can easily be integrated into a `http.Client` instance through the transport.
For example, a bare-bones script has been constructed to demonstrate this:

```go
package main

import (
	"bytes"
	"fmt"
	"github.com/rhobro/wallmask/pkg/proxy"
	"io/ioutil"
	"net/http"
)

func main() {
	// change the default transport to use
	http.DefaultTransport = &http.Transport{
		Proxy: proxy.Rand(),
	}

	rsp, _ := http.Get("https://rhobro.github.io/test") // send request using proxy
	defer rsp.Body.Close()
	body, _ := ioutil.ReadAll(rsp.Body)
	fmt.Printf("Successful: %t\n", bytes.Contains(body, []byte("TEST PAGE"))) // check if it has been received properly
}
```

This works because the proxy function is called by the `http.Transport` every time a request is sent.
Usually, `http.ProxyURL("your proxy address")` would have plugged into the transport. Instead, `proxy.Rand` returns a
function of the same signature to return a different proxy each time.