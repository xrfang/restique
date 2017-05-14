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
	conn   *sql.DB
}

var dsns map[string]dsInfo

func LoadDSNs() {
	f, err := os.Open(rc.DSN_PATH)
	assert(err)
	defer f.Close()
	dec := json.NewDecoder(f)
	assert(dec.Decode(&dsns))
}
