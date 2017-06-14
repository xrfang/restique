package main

import (
	"database/sql"
	"net/http"
	"net/url"
)

type sqlExecRes struct {
	LastInsertId int64 `json:"last_insert_id,omitempty"`
	RowsAffected int64 `json:"rows_affected"`
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
	qr, err := ds.conn.Exec(qry)
	assert(err)
	var xr sqlExecRes
	lid, err := qr.LastInsertId()
	if err == nil {
		xr.LastInsertId = lid
	}
	ra, err := qr.RowsAffected()
	if err == nil {
		xr.RowsAffected = ra
	}
	return xr
}
