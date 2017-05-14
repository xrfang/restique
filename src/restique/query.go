package main

import (
	"database/sql"
	"net/http"
	"net/url"
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
	rows, err := ds.conn.Query(qry)
	assert(err)

	cols, err := rows.Columns()
	assert(err)
	raw := make([][]byte, len(cols))
	ptr := make([]interface{}, len(cols))
	for i, _ := range raw {
		ptr[i] = &raw[i]
	}

	var recs []map[string]string
	RangeRows(rows, func() {
		assert(rows.Scan(ptr...))
		rec := map[string]string{}
		for i, r := range raw {
			if r == nil {
				rec[cols[i]] = "\\N"
			} else {
				rec[cols[i]] = string(r)
			}
		}
		recs = append(recs, rec)
	})
	return recs
}
