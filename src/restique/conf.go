package main

import (
	"os"
	"path"
	"strings"

	"github.com/xrfang/go-conf"
)

type restiqueConf struct {
	CLIENT_CIDRS  string
	SERVICE_PORT  string
	TLS_CERT      string
	TLS_PKEY      string
	AUTH_PATH     string
	DSN_PATH      string
	HIST_PATH     string
	HIST_ENTRIES  int
	OTP_ISSUER    string
	OTP_TIMEOUT   uint
	OTP_DIGITS    int
	IDLE_TIMEOUT  int
	SESSION_LIFE  int
	READ_TIMEOUT  int
	WRITE_TIMEOUT int
	QUERY_TIMEOUT int
	QUERY_MAXROWS int
	LOG_PATH      string
	LOG_ROTATE    int
	OPEN_HATEOAS  bool
	DB_TAG        string
	PID_FILE      string
}

var rc restiqueConf

func parseConfig(fn string) {
	rc.READ_TIMEOUT = 60
	rc.WRITE_TIMEOUT = 60
	rc.SERVICE_PORT = "32779"
	rc.IDLE_TIMEOUT = 300
	rc.SESSION_LIFE = 3600
	rc.OTP_DIGITS = 6
	rc.OTP_ISSUER = "restique"
	rc.OTP_TIMEOUT = 30
	rc.HIST_ENTRIES = 10
	rc.DB_TAG = "[DB]"
	rc.PID_FILE = "./restique.pid"
	rc.AUTH_PATH = "./restique_auth.json"
	rc.DSN_PATH = "./restique_dsns.json"
	rc.LOG_PATH = "./logs"
	rc.HIST_PATH = "./history"
	if fn != "" {
		assert(conf.ParseFile(fn, &rc))
	}
	if rc.IDLE_TIMEOUT > 86400 {
		rc.IDLE_TIMEOUT = 86400
	}
	if rc.SESSION_LIFE > 864000 {
		rc.SESSION_LIFE = 864000
	}
	rc.CLIENT_CIDRS = strings.TrimSpace(rc.CLIENT_CIDRS)
	if len(rc.CLIENT_CIDRS) > 0 {
		allowed_cidrs = strings.Split(rc.CLIENT_CIDRS, ",")
	}
	assert(os.MkdirAll(path.Dir(rc.AUTH_PATH), 0755))
	assert(os.MkdirAll(path.Dir(rc.DSN_PATH), 0755))
	assert(os.MkdirAll(rc.LOG_PATH, 0755))
	assert(os.MkdirAll(rc.HIST_PATH, 0755))
}
