package proxy

import "wallmask/pkg/core"

var queue = make(chan core.Proxy, 25)
