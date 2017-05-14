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
	user := flag.String("user", "", "setup/modify user account")
	pass := flag.String("pass", "", "password (optional, used with -user)")
	flag.Parse()
	if *conf == "" {
		//TODO: log errors.
		fmt.Println("missing configuration file (try -help)")
		return
	}
	rc = parseConfig(*conf)
	if *user != "" {
		SetAuth(*user, *pass)
		return
	}
	LoadAuthDb()
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler(help))
	mux.HandleFunc("/version", handler(version))
	mux.HandleFunc("/login", handler(login))
	svr := http.Server{
		Addr:         ":" + rc.SERVICE_PORT,
		Handler:      mux,
		ReadTimeout:  time.Duration(rc.READ_TIMEOUT) * time.Second,
		WriteTimeout: time.Duration(rc.WRITE_TIMEOUT) * time.Second,
	}
	if rc.TLS_CERT == "" || rc.TLS_PKEY == "" {
		assert(svr.ListenAndServe())
	} else {
		assert(svr.ListenAndServeTLS(rc.TLS_CERT, rc.TLS_PKEY))
	}
}
