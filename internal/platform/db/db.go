package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rhobro/goutils/pkg/coll"
	"github.com/rhobro/goutils/pkg/services/cfgcat"
	"github.com/rhobro/goutils/pkg/services/sentree"
	"log"
)

// Concurrency-safe pgxpool.Pool instead of pgx.Conn
var db *pgxpool.Pool
var connected bool

var reqTables = map[string]string{
	"proxies": `
		CREATE TABLE proxies (
    		id SMALLSERIAL PRIMARY KEY,
			protocol TEXT NOT NULL,
    		ipv4 TEXT NOT NULL,
    		port INT NOT NULL,
    		lastTested TIMESTAMP,
    		working BOOL NOT NULL
		);`,
}

func Connect(verbose bool) {
	// Connect to db
	var err error
	db, err = pgxpool.Connect(context.Background(), cfgcat.C.GetStringValue("dbURL", "", nil))
	if err != nil {
		log.Fatal(fmt.Errorf("{db} open connection to db: %s\n", err))
	}

	// List tables
	var tables []string
	rs := Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public';`)
	for rs.Next() {
		var tbl string
		err := rs.Scan(&tbl)
		if err != nil {
			sentree.FatalCaptureErr(fmt.Errorf("scan db table rows: %s", err))
		}
		tables = append(tables, tbl)
	}

	// Add table if doesn't exist
	for t, q := range reqTables {
		if !coll.ContainsStr(tables, t) {
			// Create table if does not exist
			Exec(q)
		}
	}

	if verbose {
		log.Print("{db} connected")
	}
	connected = true
}

func Close() {
	db.Close()
}

// For executing SQL in something like an INSERT or UPDATE statement without returning any rows.
func Exec(sql string, args ...interface{}) {
	if !connected {
		Connect(true)
	}

	rs, err := db.Query(context.Background(), sql, args...)
	if err != nil {
		sentree.C.CaptureException(err, nil, nil)
		log.Printf("{db} exec query %s: %s", sql, err)
	}
	rs.Close()
}

// Utility interface for db.Query
//
// Query executes sql with args. If there is an error the returned Rows will be returned in an error state. So it is
// allowed to ignore the error returned from Query and handle it in Rows.
//
// For extra control over how the query is executed, the types pgx.QuerySimpleProtocol, pgx.QueryResultFormats, and
// QueryResultFormatsByOID may be used as the first args to control exactly how the query is executed. This is rarely
// needed. See the documentation for those types for details.
func Query(sql string, args ...interface{}) pgx.Rows {
	if !connected {
		Connect(true)
	}

	rs, err := db.Query(context.Background(), sql, args...)
	if err != nil {
		sentree.C.CaptureException(err, nil, nil)
		log.Printf("{db} query %s: %s", sql, err)
	}
	return rs
}

// Utility interface for db.QueryFunc
//
// QueryFunc executes sql with args. For each row returned by the query the values will scanned into the elements of
// scans and f will be called. If any row fails to scan or f returns an error the query will be aborted and the error
// will be returned.
/*func QueryFunc(sql string, args []interface{}, scans []interface{}, f func(row pgx.QueryFuncRow) error) pgconn.CommandTag {
	cmd, err := db.QueryFunc(context.Background(), sql, args, scans, f)
	if err != nil {
		sentree.C.CaptureException(err, nil, nil)
		log.Printf("{db} query func %s: %s", sql, err)
	}
	return cmd
}*/

// Utility interface for db.QueryRow
//
// QueryRow is a convenience wrapper over Query. Any error that occurs while
// querying is deferred until calling Scan on the returned Row. That Row will
// error with ErrNoRows if no rows are returned.
func QueryRow(sql string, args ...interface{}) pgx.Row {
	if !connected {
		Connect(true)
	}

	return db.QueryRow(context.Background(), sql, args...)
}
