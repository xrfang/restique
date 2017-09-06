package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func doexe(conn *sql.DB, args url.Values) (queryResults, float64) {
	qry := args.Get("sql")
	ctx, cf := context.WithTimeout(context.Background(), time.Duration(rc.EXEC_TIMEOUT)*time.Second)
	defer cf()
	start := time.Now()
	qr, err := conn.ExecContext(ctx, qry)
	elapsed := time.Since(start).Seconds()
	assert(err)
	var LastInsertId, RowsAffected string
	lid, err := qr.LastInsertId()
	if err == nil {
		LastInsertId = fmt.Sprintf("%d", lid)
	} else {
		LastInsertId = err.Error()
	}
	ra, err := qr.RowsAffected()
	if err == nil {
		RowsAffected = fmt.Sprintf("%d", ra)
	} else {
		RowsAffected = err.Error()
	}
	return queryResults{
		queryResult{
			"last_insert_id": LastInsertId,
			"rows_affected":  RowsAffected,
		},
	}, elapsed
}

func exec(args url.Values) (res interface{}) {
	use := args.Get("use")
	qry := args.Get("sql")
	if use == "" || qry == "" {
		return httpError{
			Code: http.StatusSeeOther,
			Mesg: "/uisql?action=exec&use=" + use,
		}
	}
	ds, ok := dsns[use]
	if !ok {
		return httpError{
			Code: http.StatusNotFound,
			Mesg: "[use] is not a valid data source",
		}
	}
	var (
		dss []dsInfo
		els float64
		rec queryResults
		err error
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
		data, elapsed := doexe(conn, args)
		els += elapsed
		for _, d := range data {
			if len(dss) > 1 {
				d[rc.DB_TAG] = ds.Name
			}
			rec = append(rec, d)
		}
	}
	summary := fmt.Sprintf("Executed statement in %fs", els)
	args.Set("RESTIQUE_SUMMARY", summary)
	return rec
}
