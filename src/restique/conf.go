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
	USER_DATABASE string
	IDLE_TIMEOUT  int
	READ_TIMEOUT  int
	WRITE_TIMEOUT int
}

func parseConfig(fn string) (rc restiqueConf) {
	rc.READ_TIMEOUT = 60
	rc.WRITE_TIMEOUT = 60
	rc.SERVICE_PORT = "32779"
	rc.IDLE_TIMEOUT = 600
	assert(conf.ParseFile(fn, &rc))
	if rc.USER_DATABASE == "" {
		rc.USER_DATABASE = "./" + self + ".json"
	}
	if rc.IDLE_TIMEOUT > 86400 {
		rc.IDLE_TIMEOUT = 86400
	}
	assert(os.MkdirAll(path.Dir(rc.USER_DATABASE), 0755))
	return
}
