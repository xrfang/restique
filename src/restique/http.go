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
	return he.Mesg
}

type session struct {
	id string
	ip string
	un string
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
	ss.Lock()
	ss.s[sid] = session{
		id: sid,
		ip: ip,
		un: r.Form.Get("name"),
		ex: now.Add(idle),
		lt: now.Add(life),
	}
	ss.Unlock()
	return sid
}

func (ss sessionStore) Get(r *http.Request) (session, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		return session{}, false
	}
	ss.Lock()
	defer ss.Unlock()
	s, ok := ss.s[c.Value]
	if !ok {
		return session{}, false
	}
	ip := strings.Split(r.RemoteAddr, ":")[0]
	if s.ex.Before(time.Now()) || s.ip != ip {
		return session{}, false
	}
	ex := time.Now().Add(time.Duration(rc.IDLE_TIMEOUT) * time.Second)
	if ex.After(s.lt) {
		s.ex = s.lt
	} else {
		s.ex = ex
	}
	ss.s[c.Value] = s
	return s, true
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
	_, ok := ss.Get(r)
	return ok
}

func handler(proc func(url.Values) interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var args url.Values
		var out bytes.Buffer
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
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
				if code >= 300 && code < 400 {
					http.Redirect(w, r, data, code)
				} else {
					http.Error(w, data, code)
				}
			}
			if strings.Contains(args.Get("REQUEST_URL_PATH"), "login") {
				delete(args, "code")
				delete(args, "pass")
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
		r.ParseForm()
		args = r.Form
		args.Add("REQUEST_URL_PATH", r.URL.Path)
		data := proc(args)
		if e, ok := data.(httpError); ok {
			if r.URL.Path == "/loginui" {
				panic(httpError{
					Code: http.StatusFound,
					Mesg: fmt.Sprintf("/uilgn?name=%s&err=%s", args.Get("name"),
						e.Error()),
				})
			}
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
				panic(httpError{
					Code: http.StatusSeeOther,
					Mesg: "/query?" + sid,
				})
			}
			data = map[string]string{"session": sid}
		}
		mw := io.MultiWriter(&out, w)
		enc := json.NewEncoder(mw)
		enc.SetIndent("", "    ")
		assert(enc.Encode(data))
	}
}
