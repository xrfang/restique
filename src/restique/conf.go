package main

import (
	"github.com/xrfang/go-conf"
)

type restiqueConf struct {
	CLIENT_CIDRS  string
	SERVICE_PORT  string
	TLS_CERT      string
	TLS_PKEY      string
	AUTH_PATH     string
	DSN_PATH      string
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
}

func parseConfig(fn string) (rc restiqueConf) {
	rc.READ_TIMEOUT = 60
	rc.WRITE_TIMEOUT = 60
	rc.SERVICE_PORT = "32779"
	rc.IDLE_TIMEOUT = 300
	rc.SESSION_LIFE = 3600
	rc.OTP_DIGITS = 6
	rc.OTP_ISSUER = "restique"
	rc.OTP_TIMEOUT = 30
	if fn != "" {
		assert(conf.ParseFile(fn, &rc))
	}
	if rc.AUTH_PATH == "" {
		rc.AUTH_PATH = "./restique_auth.json"
	}
	if rc.DSN_PATH == "" {
		rc.DSN_PATH = "./restique_dsns.json"
	}
	if rc.LOG_PATH == "" {
		rc.LOG_PATH = "./logs"
	}
	return
}
