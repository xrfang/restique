package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

var rc restiqueConf

func main() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("ERROR: %v\n", e)
		}
	}()
	conf := flag.String("conf", "", "configuration file")
	flag.Parse()
	if *conf == "" {
		//TODO: log errors.
		return
	}
	rc = parseConfig(*conf)
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler(help))
	mux.HandleFunc("/version", handler(version))
	svr := http.Server{
		Addr:         ":" + rc.SERVICE_PORT,
		Handler:      mux,
		ReadTimeout:  time.Duration(rc.READ_TIMEOUT) * time.Second,
		WriteTimeout: time.Duration(rc.WRITE_TIMEOUT) * time.Second,
	}
	if rc.tls {
		assert(svr.ListenAndServeTLS(rc.TLS_CERT, rc.TLS_PKEY))
	} else {
		assert(svr.ListenAndServe())
	}
}
