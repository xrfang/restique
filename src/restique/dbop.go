package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type dsInfo struct {
	Driver string
	Dsn    string
	Memo   string
	Name   string
}

var dsns map[string]dsInfo

func LoadDSNs() {
	f, err := os.Open(rc.DSN_PATH)
	assert(err)
	defer f.Close()
	dec := json.NewDecoder(f)
	assert(dec.Decode(&dsns))
}

func RangeRows(rows *sql.Rows, proc func()) {
	defer func() {
		if e := recover(); e != nil {
			rows.Close()
			panic(e)
		}
	}()
	for rows.Next() {
		proc()
	}
	assert(rows.Err())
}

func ExpandMultiDSN(ds dsInfo) ([]dsInfo, error) {
	var dss []dsInfo
	for _, n := range strings.Split(ds.Dsn, ",") {
		s, ok := dsns[n]
		if !ok {
			return nil, httpError{
				Code: http.StatusInternalServerError,
				Mesg: "Unknown DSN in [multi]: " + n,
			}
		}
		if s.Driver == "[multi]" {
			return nil, httpError{
				Code: http.StatusInternalServerError,
				Mesg: "[multi] data source cannot be nested",
			}
		}
		s.Name = n
		dss = append(dss, s)
	}
	return dss, nil
}
