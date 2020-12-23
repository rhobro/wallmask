package db

import (
	"database/sql"
	"fmt"
	"github.com/Bytesimal/goutils/pkg/coll"
	_ "github.com/ibmdb/go_ibm_db"
	"log"
	"sync"
)

const dsn = "DATABASE=BLUDB;HOSTNAME=dashdb-txn-sbox-yp-lon02-04.services.eu-gb.bluemix.net;PORT=50001;PROTOCOL=TCPIP;UID=lzx36405;PWD=8k11s2d98lhk81^p;Security=SSL;"

var DB *sql.DB
var connectNotice sync.Once

var reqTables = map[string]string{
	"PRX_PROXIES": `
		CREATE TABLE prx_proxies (
			id INT PRIMARY KEY NOT NULL,
			ipv4 VARCHAR(15) NOT NULL,
			port INT,
    		lastTested INT
		);`,
}

func init() {
	// Connect to db
	var err error
	DB, err = sql.Open("go_ibm_db", dsn)
	if err != nil {
		log.Fatalf("unable to open connection to DB: %s\n", err)
	}

	// Check if req tables
	st, err := DB.Prepare(`
	SELECT table_name
	FROM sysibm.tables
	WHERE table_schema = 'LZX36405';`)
	if err != nil {
		log.Fatalf("unable to access scema list: %s\n", err)
	}
	defer st.Close()

	rs, err := Exec(st)
	if err != nil {
		log.Fatalf("unable to access scema list exec: %s\n", err)
	}
	defer rs.Close()

	var tables []string
	for rs.Next() {
		var nS string
		err := rs.Scan(&nS)
		if err != nil {
			tables = append(tables)
		}
		tables = append(tables, nS)
	}

	for t, q := range reqTables {
		if !coll.ContainsStr(tables, t) {
			// Create table if does not exist
			st, err := DB.Prepare(q)
			if err != nil {
				log.Fatalf("can't create table %s with q %s: %s", t, q, err)
			}

			rs, err := Exec(st)
			if err != nil {
				log.Fatalf("{proxy} {db} Unable to create table %s: %s\n", t, err)
			}
			st.Close()
			rs.Close()
		}
	}
}

func Exec(st *sql.Stmt) (s *sql.Rows, err error) {
	connectNotice.Do(func() {
		log.Println("{proxy} {db} Connected")
	})

	err = fmt.Errorf("tmp")
	for i := 0; i < 10; i++ {
		if err != nil {
			s, err = st.Query()
		} else {
			break
		}
	}
	return
}
