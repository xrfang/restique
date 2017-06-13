package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type httpError struct {
	Code int
	Mesg string
}

func (he httpError) Error() string {
	return fmt.Sprintf("HTTP/%d: %s", he.Code, he.Mesg)
}

func val(args url.Values, key string) string {
	a := args[key]
	if len(a) == 0 {
		return ""
	}
	return a[0]
}

type session struct {
	id string
	ip string
	ex time.Time
	lt time.Time
}

type sessionStore struct {
	s map[string]session
	sync.RWMutex
}

var sessions sessionStore

func init() {
	sessions.s = make(map[string]session)
	go func() {
		for {
			<-time.After(time.Minute)
			for k, v := range sessions.s {
				if time.Now().After(v.ex) {
					sessions.Lock()
					delete(sessions.s, k)
					sessions.Unlock()
				}
			}
		}
	}()
}

func (ss sessionStore) NewSession(r *http.Request) string {
	charmap := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	N := 16
	b := make([]byte, N)
	rand.Read(b)
	for i := 0; i < N; i++ {
		r := b[i] % 62
		b[i] = charmap[r]
	}
	sid := string(b)
	ip := strings.Split(r.RemoteAddr, ":")[0]
	now := time.Now()
	idle := time.Duration(rc.IDLE_TIMEOUT) * time.Second
	life := time.Duration(rc.SESSION_LIFE) * time.Second
	ss.s[sid] = session{
		id: sid,
		ip: ip,
		ex: now.Add(idle),
		lt: now.Add(life),
	}
	return sid
}

func (ss sessionStore) SessionOK(r *http.Request) bool {
	switch r.URL.Path {
	case "/", "/uilgn", "/login", "/loginui":
		return true
	case "/api", "/version":
		if rc.OPEN_HATEOAS {
			return true
		}
	}
	c, err := r.Cookie("session")
	if err != nil {
		return false
	}
	s, ok := ss.s[c.Value]
	if !ok {
		return false
	}
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if s.ex.Before(time.Now()) || s.ip != ip {
		return false
	}
	ex := time.Now().Add(time.Duration(rc.IDLE_TIMEOUT) * time.Second)
	if ex.After(s.lt) {
		s.ex = s.lt
	} else {
		s.ex = ex
	}
	ss.s[c.Value] = s
	return true
}

func handler(proc func(url.Values) interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args url.Values
		var out bytes.Buffer
		requestTime := time.Now()
		defer func() {
			code := http.StatusOK
			data := out.String()
			if e := recover(); e != nil {
				switch e.(type) {
				case httpError:
					code = e.(httpError).Code
					data = e.(httpError).Mesg
				default:
					code = http.StatusInternalServerError
					data = e.(error).Error()
				}
				if code == http.StatusSeeOther {
					http.Redirect(w, r, data, code)
				} else {
					http.Error(w, data, code)
				}
			}
			delete(args, "REQUEST_URL_PATH")
			lms <- logMessage{
				Client:   r.RemoteAddr,
				Time:     requestTime,
				Duration: time.Since(requestTime).Seconds(),
				Request:  r.URL.Path,
				Params:   args,
				Cookie:   r.Cookies(),
				Code:     code,
				Reply:    data,
			}
		}()
		if AccessDenied(r) {
			panic(httpError{Code: http.StatusForbidden, Mesg: "access denied"})
		}
		if !sessions.SessionOK(r) {
			panic(httpError{Code: http.StatusSeeOther, Mesg: "/uilgn"})
		}
		if r.Method == "POST" || r.Method == "PUT" {
			r.ParseForm()
			args = r.Form
		} else {
			args = r.URL.Query()
		}
		args["REQUEST_URL_PATH"] = []string{r.URL.Path}
		data := proc(args)
		if e, ok := data.(httpError); ok {
			panic(e)
		}
		if strings.HasPrefix(r.URL.Path, "/login") {
			sid := sessions.NewSession(r)
			http.SetCookie(w, &http.Cookie{
				Name:    "session",
				Value:   sid,
				Path:    "/",
				Expires: time.Now().Add(24 * time.Hour),
			})
			if r.URL.Path == "/loginui" {
				panic(httpError{Code: http.StatusSeeOther, Mesg: "/query"})
			}
			data = map[string]string{"session": sid}
		}
		mw := io.MultiWriter(&out, w)
		enc := json.NewEncoder(mw)
		enc.SetIndent("", "    ")
		assert(enc.Encode(data))
	}
}
