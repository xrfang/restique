package main

import (
	"github.com/xrfang/go-conf"
)

type restiqueConf struct {
	SERVICE_PORT  string
	TLS_CERT      string
	TLS_PKEY      string
	READ_TIMEOUT  int
	WRITE_TIMEOUT int
	tls           bool
}

func parseConfig(fn string) (rc restiqueConf) {
	rc.READ_TIMEOUT = 60
	rc.WRITE_TIMEOUT = 60
	assert(conf.ParseFile(fn, &rc))
	tls := rc.TLS_CERT != "" && rc.TLS_PKEY != ""
	if rc.SERVICE_PORT == "" {
		if tls {
			rc.SERVICE_PORT = "443"
		} else {
			rc.SERVICE_PORT = "80"
		}
	}
	return
}
