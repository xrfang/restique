package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"gopass"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
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
	dsn_path := flag.String("dsn-path", "", "DSN info file path")
	dsn_init := flag.Bool("dsn-init", false, "generate DSN info file template")
	service_port := flag.String("service-port", "", "HTTP(S) service port")
	tls_cert := flag.String("tls-cert", "", "TLS certification")
	tls_pkey := flag.String("tls-pkey", "", "TLS private key")
	idle_timeout := flag.Int("idle-timeout", 0, "session idle timeout")
	session_life := flag.Int("session-life", 0, "max session lifetime")
	read_timeout := flag.Int("read-timeout", 0, "timeout for HTTP request")
	qry_timeout := flag.Int("query-timeout", 0, "timeout for executing queries")
	qry_maxrows := flag.Int("query-maxrows", 0, "maximum rows to return for queries")
	write_timeout := flag.Int("write-timeout", 0, "timeout for HTTP reply")
	otp_digits := flag.Int("otp-digits", 0, "OTP code length (6~8 recommended)")
	otp_issuer := flag.String("otp-issuer", "", "OTP issuer for information purpose")
	otp_timeout := flag.Uint("otp-timeout", 0, "OTP code lifetime (30~60 recommended)")
	client_cidrs := flag.String("client-cidrs", "", "Access control via IP range")
	ver := flag.Bool("version", false, "show version info")
	log_path := flag.String("log-path", "", "directory to save log files")
	log_rotate := flag.Int("log-rotate", 0, "days to keep log files (0=keep forever)")
	hateoas := flag.Bool("hateoas", false, "show API info without authentication")
	flag.Parse()

	if *ver {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")
		enc.Encode(version(nil))
		return
	}

	rc = parseConfig(*conf)
	if *dsn_path != "" {
		rc.DSN_PATH = *dsn_path
	}
	assert(os.MkdirAll(path.Dir(rc.DSN_PATH), 0755))

	if *dsn_init {
		fmt.Printf("DSN configuration: %s\n", rc.DSN_PATH)
		_, err := os.Stat(rc.DSN_PATH)
		if err == nil {
			fmt.Println("file already exists, abort.")
			return
		}
		f, err := os.Create(rc.DSN_PATH)
		assert(err)
		defer f.Close()
		enc := json.NewEncoder(f)
		enc.SetIndent("", "    ")
		assert(enc.Encode(map[string]map[string]string{
			"sample_conn": map[string]string{
				"driver": "mysql",
				"dsn":    "connection string",
				"memo":   "a placeholder connection",
			},
		}))
		fmt.Println("template created.")
		return
	}

	if *auth_path != "" {
		rc.AUTH_PATH = *auth_path
	}
	assert(os.MkdirAll(path.Dir(rc.AUTH_PATH), 0755))

	if *log_path != "" {
		rc.LOG_PATH = *log_path
	}
	assert(os.MkdirAll(rc.LOG_PATH, 0755))

	LoadAuthDb()
	if *user != "" {
		pswd := *pass
		if *pass == "" {
			var err error
			prompt := fmt.Sprintf("Enter password for %s: ", *user)
			pswd, err = gopass.GetPass(prompt)
			assert(err)
			pswd2, err := gopass.GetPass("Enter password again: ")
			assert(err)
			if pswd != pswd2 {
				fmt.Println("ERROR: password mismatch, aborted.")
				return
			}
		}
		SetAuth(*user, pswd)
		return
	}
	LoadDSNs()

	if *service_port != "" {
		rc.SERVICE_PORT = *service_port
	}
	if *tls_cert != "" {
		rc.TLS_CERT = *tls_cert
	}
	if *tls_pkey != "" {
		rc.TLS_PKEY = *tls_pkey
	}
	if *log_rotate > 0 {
		rc.LOG_ROTATE = *log_rotate
	}
	if *idle_timeout > 0 {
		rc.IDLE_TIMEOUT = *idle_timeout
	}
	if rc.IDLE_TIMEOUT > 3600 {
		rc.IDLE_TIMEOUT = 3600
	}
	if *session_life > 0 {
		rc.SESSION_LIFE = *session_life
	}
	if rc.SESSION_LIFE > 86400 {
		rc.SESSION_LIFE = 86400
	}
	if *read_timeout > 0 {
		rc.READ_TIMEOUT = *read_timeout
	}
	if *write_timeout > 0 {
		rc.WRITE_TIMEOUT = *write_timeout
	}
	if *qry_timeout > 0 {
		rc.QUERY_TIMEOUT = *qry_timeout
	}
	if *qry_maxrows > 0 {
		rc.QUERY_MAXROWS = *qry_maxrows
	}
	if *otp_digits > 0 {
		rc.OTP_DIGITS = *otp_digits
	}
	if *otp_issuer != "" {
		rc.OTP_ISSUER = *otp_issuer
	}
	if *otp_timeout > 0 {
		rc.OTP_TIMEOUT = *otp_timeout
	}
	if *client_cidrs != "" {
		rc.CLIENT_CIDRS = *client_cidrs
	}
	rc.CLIENT_CIDRS = strings.TrimSpace(rc.CLIENT_CIDRS)
	if len(rc.CLIENT_CIDRS) > 0 {
		allowed_cidrs = strings.Split(rc.CLIENT_CIDRS, ",")
	}
	if *hateoas {
		rc.OPEN_HATEOAS = true
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler(home))
	mux.HandleFunc("/api", handler(help))
	mux.HandleFunc("/version", handler(version))
	mux.HandleFunc("/login", handler(login))
	mux.HandleFunc("/loginui", handler(login))
	mux.HandleFunc("/query", handler(query))
	mux.HandleFunc("/exec", handler(exec))
	mux.HandleFunc("/conns", handler(conns))
	mux.HandleFunc("/uilgn", uiLgn)
	mux.HandleFunc("/uisql", uiSql)

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGHUP)
	go func() {
		for {
			switch <-sigch {
			case syscall.SIGHUP:
				LoadAuthDb()
				LoadDSNs()
			}
		}
	}()

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
