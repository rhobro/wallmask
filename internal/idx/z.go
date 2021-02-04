package idx

import "log"

func Index() {
	// launch db testers
	go dbTest(true, ASC, -1)    // testing working proxies
	go dbTest(false, DESC, 500) // testing first few dead proxies
	go dbTest(false, DESC, -1)  // testing dead proxies
	for i := 0; i < nTestWorkers; i++ {
		go testWorker()
	}

	// launch idx scheduler
	go scheduler()
	log.Print("{proxy} initialized")
}
