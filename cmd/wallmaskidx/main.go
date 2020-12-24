package main

import (
	"bufio"
	"github.com/Bytesimal/goutils/pkg/fileio"
	_ "github.com/ibmdb/go_ibm_db"
	"os"
	"wallmask/internal/idx"
)

func init() {
	fileio.Init("", "wallmaskidx")
}

/*func initTest() {
	urll, _ := url.Parse("http://localhost:9090")
	http.DefaultClient.Transport = &http.Transport{
		Proxy: http.ProxyURL(urll),
	}
}*/

func main() {
	// Start indexing
	idx.Index()
	// Wait until user clicks enter
	rd := bufio.NewScanner(os.Stdin)
	rd.Scan()
}
