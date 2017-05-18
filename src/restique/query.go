package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func query(args url.Values) (res interface{}) {
	use := val(args, "use")
	qry := val(args, "sql")
	if use == "" || qry == "" {
		return httpError{
			Code: http.StatusBadRequest,
			Mesg: "[use] or [sql] not provided",
		}
	}
	ds, ok := dsns[use]
	if !ok {
		return httpError{
			Code: http.StatusNotFound,
			Mesg: "[use] is not a valid data source",
		}
	}
	defer func() {
		if e := recover(); e != nil {
			ds.conn.Close()
			ds.conn = nil
			res = httpError{
				Code: http.StatusInternalServerError,
				Mesg: e.(error).Error(),
			}
		}
	}()
	if ds.conn == nil {
		conn, err := sql.Open(ds.Driver, ds.Dsn)
		assert(err)
		ds.conn = conn
	}
	var (
		rows        *sql.Rows
		timeoutChan <-chan time.Time
		resultChan  chan error
	)
	if rc.QUERY_TIMEOUT > 0 {
		timeoutChan = time.After(time.Duration(rc.QUERY_TIMEOUT) * time.Second)
	}
	resultChan = make(chan error)
	go func() {
		var err error
		rows, err = ds.conn.Query(qry)
		resultChan <- err
	}()
	select {
	case <-timeoutChan:
		panic(fmt.Errorf("query timeout exceeded (%d)", rc.QUERY_TIMEOUT))
	case err := <-resultChan:
		assert(err)
	}

	cols, err := rows.Columns()
	assert(err)
	raw := make([][]byte, len(cols))
	ptr := make([]interface{}, len(cols))
	for i, _ := range raw {
		ptr[i] = &raw[i]
	}
	recs := []map[string]interface{}{}
	RangeRows(rows, func() {
		assert(rows.Scan(ptr...))
		rec := map[string]interface{}{}
		for i, r := range raw {
			if r == nil {
				rec[cols[i]] = nil
			} else {
				rec[cols[i]] = string(r)
			}
		}
		recs = append(recs, rec)
		if rc.QUERY_MAXROWS > 0 && len(recs) > rc.QUERY_MAXROWS {
			panic(fmt.Errorf("at most %d rows can be fetched (try use LIMIT)",
				rc.QUERY_MAXROWS))
		}
	})
	return recs
}
