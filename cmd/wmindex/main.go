package main

import (
	"database/sql"
	"fmt"
	_ "github.com/ibmdb/go_ibm_db"
	"net/http"
	"net/url"
)

func init() {
	dbg, _ := url.Parse("http://localhost:9090")
	http.DefaultClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(dbg),
		},
	}
}

const dsn = "DATABASE=BLUDB;HOSTNAME=dashdb-txn-sbox-yp-lon02-04.services.eu-gb.bluemix.net;PORT=50001;PROTOCOL=TCPIP;UID=lzx36405;PWD=8k11s2d98lhk81^p;Security=SSL;"

var DB *sql.DB

func main() {
	var err error
	DB, err = sql.Open("go_ibm_db", dsn)
	if err != nil {
		panic(fmt.Sprintf("Unable to open connection to DB: %s\n", err))
	}
}
