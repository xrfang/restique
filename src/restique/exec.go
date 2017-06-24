package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type sqlExecRes struct {
	LastInsertId string `json:"last_insert_id"`
	RowsAffected string `json:"rows_affected"`
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
	start := time.Now()
	qr, err := ds.conn.Exec(qry)
	elapsed := time.Since(start).Seconds()
	assert(err)
	var xr sqlExecRes
	lid, err := qr.LastInsertId()
	if err == nil {
		xr.LastInsertId = fmt.Sprintf("%d", lid)
	} else {
		xr.LastInsertId = err.Error()
	}
	ra, err := qr.RowsAffected()
	if err == nil {
		xr.RowsAffected = fmt.Sprintf("%d", ra)
	} else {
		xr.RowsAffected = err.Error()
	}
	summary := fmt.Sprintf("Executed statement in %fs", elapsed)
	args.Set("RESTIQUE_SUMMARY", summary)
	return xr
}
