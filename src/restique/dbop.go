package main

import (
	"database/sql"
	"encoding/json"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

type dsInfo struct {
	Driver string
	Dsn    string
	Memo   string
	//conn   *sql.DB
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
