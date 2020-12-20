package db

import (
	"database/sql"
	"fmt"
	_ "github.com/ibmdb/go_ibm_db"
)

const dsn = "DATABASE=BLUDB;HOSTNAME=dashdb-txn-sbox-yp-lon02-04.services.eu-gb.bluemix.net;PORT=50001;PROTOCOL=TCPIP;UID=lzx36405;PWD=8k11s2d98lhk81^p;Security=SSL;"

var DB *sql.DB

func init() {
	var err error
	DB, err = sql.Open("go_ibm_db", dsn)
	if err != nil {
		panic(fmt.Sprintf("Unable to open connection to DB: %s\n", err))
	}
}

func Exec(st *sql.Stmt) (s *sql.Rows, err error) {
	s, err = st.Query()
	return
}
