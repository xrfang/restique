package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func doqry(conn *sql.DB, args url.Values) (queryResults, float64, float64) {
	var (
		rows        *sql.Rows
		timeoutChan <-chan time.Time
		resultChan  chan error
		tq, tf      float64
	)
	qry := args.Get("sql")
	if rc.QUERY_TIMEOUT > 0 {
		timeoutChan = time.After(time.Duration(rc.QUERY_TIMEOUT) * time.Second)
	}
	resultChan = make(chan error)
	start := time.Now()
	go func() {
		var err error
		rows, err = conn.Query(qry)
		resultChan <- err
	}()
	select {
	case <-timeoutChan:
		panic(fmt.Errorf("query timeout exceeded (%d)", rc.QUERY_TIMEOUT))
	case err := <-resultChan:
		assert(err)
		tq = time.Since(start).Seconds()
	}
	start = time.Now()
	cols, err := rows.Columns()
	assert(err)
	raw := make([][]byte, len(cols))
	ptr := make([]interface{}, len(cols))
	for i, _ := range raw {
		ptr[i] = &raw[i]
	}
	recs := queryResults{}
	RangeRows(rows, func() {
		assert(rows.Scan(ptr...))
		rec := queryResult{}
		for i, r := range raw {
			if r == nil {
				rec[cols[i]] = nil
			} else {
				rec[cols[i]] = string(r)
			}
		}
		if rc.QUERY_MAXROWS > 0 && len(recs) > rc.QUERY_MAXROWS {
			args.Set("RESTIQUE_MAXROW", "1")
			return
		}
		recs = append(recs, rec)
	})
	tf = time.Since(start).Seconds()
	return recs, tq, tf
}

func query(args url.Values) (res interface{}) {
	use := args.Get("use")
	qry := args.Get("sql")
	if use == "" || qry == "" {
		return httpError{
			Code: http.StatusSeeOther,
			Mesg: "/uisql?action=query&use=" + use,
		}
	}
	ds, ok := dsns[use]
	if !ok {
		return httpError{
			Code: http.StatusInternalServerError,
			Mesg: "[use] is not a valid data source",
		}
	}
	var (
		dss      []dsInfo
		recs     queryResults
		tqs, tfs float64
		err      error
	)
	if ds.Driver == "[multi]" {
		dss, err = ExpandMultiDSN(ds)
		if err != nil {
			return err
		}
	} else {
		ds.Name = use
		dss = append(dss, ds)
	}
	defer func() {
		if e := recover(); e != nil {
			res = httpError{
				Code: http.StatusInternalServerError,
				Mesg: e.(error).Error(),
			}
		}
	}()
	for _, ds := range dss {
		conn, err := sql.Open(ds.Driver, ds.Dsn)
		assert(err)
		data, tq, tf := doqry(conn, args)
		tqs += tq
		tfs += tf
		for _, d := range data {
			if len(dss) > 1 {
				d[rc.DB_TAG] = ds.Name
			}
			recs = append(recs, d)
		}
	}
	summary := ""
	if len(recs) < 2 {
		summary = fmt.Sprintf("Got %d row in %fs (query=%fs; fetch=%fs)",
			len(recs), tqs+tfs, tqs, tfs)

	} else {
		summary = fmt.Sprintf("Got %d rows in %fs (query=%fs; fetch=%fs)",
			len(recs), tqs+tfs, tqs, tfs)
	}
	args.Set("RESTIQUE_SUMMARY", summary)
	return recs
}
