package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"gopass"
	"net/http"
	"os"
	"time"

	"github.com/mdp/qrterminal"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func main() {
	defer func() {
		if e := recover(); e != nil {
			msg := trace("ERROR: %v", e)
			for _, m := range msg {
				fmt.Println(m)
			}
		}
	}()
	conf := flag.String("conf", "", "configuration file")
	user := flag.String("user", "", "setup/modify user account")
	pass := flag.String("pass", "", "password (optional, used with -user)")
	dsn_init := flag.Bool("dsn-init", false, "generate DSN info file template")
	ver := flag.Bool("version", false, "show version info")
	flag.Parse()

	if *ver {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "    ")
		enc.Encode(version(nil))
		return
	}

	parseConfig(*conf)
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
		want_otp := true
		otpkey := ""
		if pswd != "" {
			r := bufio.NewReader(os.Stdin)
			fmt.Print("Enable two-factor authentication (OTP)? [Y/n] ")
			yn, _ := r.ReadString('\n')
			if len(yn) > 0 && (yn[0] == 'N' || yn[0] == 'n') {
				want_otp = false
			}
		}
		if want_otp {
			gopts := totp.GenerateOpts{
				AccountName: *user,
				Digits:      otp.Digits(rc.OTP_DIGITS),
				Issuer:      rc.OTP_ISSUER,
				Period:      rc.OTP_TIMEOUT,
			}
			key, err := totp.Generate(gopts)
			assert(err)
			qrterminal.Generate(key.String(), qrterminal.L, os.Stdout)
			otpkey = key.Secret()
		}
		SetAuth(*user, pswd, otpkey)
		reloadConfig()
		return
	}
	LoadDSNs()
	handleSignals()
	savePid()
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
