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
	auth_path := flag.String("auth-path", "", "authentication file path")
	service_port := flag.String("service-port", "", "HTTP(S) service port")
	tls_cert := flag.String("tls-cert", "", "TLS certification")
	tls_pkey := flag.String("tls-pkey", "", "TLS private key")
	idle_timeout := flag.Int("idle-timeout", 0, "session idle timeout")
	read_timeout := flag.Int("read-timeout", 0, "timeout for HTTP request")
	write_timeout := flag.Int("write-timeout", 0, "timeout for HTTP reply")
	flag.Parse()

	rc = parseConfig(*conf)
	if *auth_path != "" {
		rc.AUTH_PATH = *auth_path
	}
	if *user != "" {
		SetAuth(*user, *pass)
		return
	}
	LoadAuthDb()

	if *service_port != "" {
		rc.SERVICE_PORT = *service_port
	}
	if *tls_cert != "" {
		rc.TLS_CERT = *tls_cert
	}
	if *tls_pkey != "" {
		rc.TLS_PKEY = *tls_pkey
	}
	if *idle_timeout > 0 {
		rc.IDLE_TIMEOUT = *idle_timeout
	}
	if *read_timeout > 0 {
		rc.READ_TIMEOUT = *read_timeout
	}
	if *write_timeout > 0 {
		rc.WRITE_TIMEOUT = *write_timeout
	}

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
