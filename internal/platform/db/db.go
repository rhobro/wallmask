package db

import (
	"context"
	"github.com/Bytesimal/goutils/pkg/coll"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
)

// DSN to connect to test CockroachDB
const dsn = "postgres://root@localhost:26257/wallmask"

// Concurrency-safe pgxpool.Pool instead of pgx.Conn
var db *pgxpool.Pool

var reqTables = map[string]string{
	"proxies": `
		CREATE TABLE proxies (
    		id SMALLSERIAL PRIMARY KEY,
    		ipv4 TEXT NOT NULL,
    		port INT,
    		lastTested TIMESTAMP,
    		working BOOL NOT NULL
		);`,
}

func init() {
	// Connect to db
	var err error
	db, err = pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("{proxy} {db} open connection to db: %s\n", err)
	}
	log.Println("{proxy} {db} Connected")

	// List tables
	var tables []string
	rs, err := Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public';`)
	if err != nil {
		log.Fatalf("querying db tables: %s", err)
	}
	for rs.Next() {
		var tbl string
		err := rs.Scan(&tbl)
		if err != nil {
			log.Fatalf("scan db table rows: %s", err)
		}
		tables = append(tables, tbl)
	}

	// Add table if doesn't exist
	for t, q := range reqTables {
		if !coll.ContainsStr(tables, t) {
			// Create table if does not exist
			err := Exec(q)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// For executing SQL in something like an INSERT or UPDATE statement without returning any rows.
func Exec(sql string, args ...interface{}) (err error) {
	rs, err := Query(sql, args...)
	if err == nil {
		rs.Close()
	}
	return
}

// Utility interface for db.Query
//
// Query executes sql with args. If there is an error the returned Rows will be returned in an error state. So it is
// allowed to ignore the error returned from Query and handle it in Rows.
//
// For extra control over how the query is executed, the types pgx.QuerySimpleProtocol, pgx.QueryResultFormats, and
// QueryResultFormatsByOID may be used as the first args to control exactly how the query is executed. This is rarely
// needed. See the documentation for those types for details.
func Query(sql string, args ...interface{}) (pgx.Rows, error) {
	return db.Query(context.Background(), sql, args...)
}

// Utility interface for db.QueryFunc
//
// QueryFunc executes sql with args. For each row returned by the query the values will scanned into the elements of
// scans and f will be called. If any row fails to scan or f returns an error the query will be aborted and the error
// will be returned.
func QueryFunc(sql string, args []interface{}, scans []interface{}, f func(row pgx.QueryFuncRow) error) (pgconn.CommandTag, error) {
	return db.QueryFunc(context.Background(), sql, args, scans, f)
}

// Utility interface for db.QueryRow
//
// QueryRow is a convenience wrapper over Query. Any error that occurs while
// querying is deferred until calling Scan on the returned Row. That Row will
// error with ErrNoRows if no rows are returned.
func QueryRow(sql string, args ...interface{}) pgx.Row {
	return db.QueryRow(context.Background(), sql, args...)
}
