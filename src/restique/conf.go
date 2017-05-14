package main

import (
	"os"
	"path"

	"github.com/xrfang/go-conf"
)

type restiqueConf struct {
	SERVICE_PORT  string
	TLS_CERT      string
	TLS_PKEY      string
	AUTH_PATH     string
	DSN_PATH      string
	IDLE_TIMEOUT  int
	READ_TIMEOUT  int
	WRITE_TIMEOUT int
}

func parseConfig(fn string) (rc restiqueConf) {
	rc.READ_TIMEOUT = 60
	rc.WRITE_TIMEOUT = 60
	rc.SERVICE_PORT = "32779"
	rc.IDLE_TIMEOUT = 600
	if fn != "" {
		assert(conf.ParseFile(fn, &rc))
	}
	if rc.AUTH_PATH == "" {
		rc.AUTH_PATH = "./restique_auth.json"
	}
	if rc.DSN_PATH == "" {
		rc.DSN_PATH = "./restique_dsns.conf"
	}
	if rc.IDLE_TIMEOUT > 86400 {
		rc.IDLE_TIMEOUT = 86400
	}
	assert(os.MkdirAll(path.Dir(rc.AUTH_PATH), 0755))
	return
}
