package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
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
	ss.s[sid] = session{
		id: sid,
		ip: ip,
		ex: time.Now().Add(time.Duration(rc.IDLE_TIMEOUT) * time.Second),
	}
	return sid
}

func (ss sessionStore) SessionOK(r *http.Request) bool {
	switch r.URL.Path {
	case "/", "/login", "/version":
		return true
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
	s.ex = time.Now().Add(time.Duration(rc.IDLE_TIMEOUT) * time.Second)
	ss.s[c.Value] = s
	return true
}

func handler(proc func(url.Values) interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				//TODO: logging?
				http.Error(w, e.(error).Error(), http.StatusInternalServerError)
			}
		}()
		if AccessDenied(r) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}
		if !sessions.SessionOK(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var args url.Values
		if r.Method == "POST" || r.Method == "PUT" {
			r.ParseForm()
			args = r.Form
		} else {
			args = r.URL.Query()
		}
		args["REQUEST_URL_PATH"] = []string{r.URL.Path}
		data := proc(args)
		if e, ok := data.(httpError); ok {
			http.Error(w, e.Error(), e.Code)
			return
		}
		if r.URL.Path == "/login" {
			sid := sessions.NewSession(r)
			http.SetCookie(w, &http.Cookie{
				Name:    "session",
				Value:   sid,
				Path:    "/",
				Expires: time.Now().Add(24 * time.Hour),
			})
			data = map[string]string{"session": sid}
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		assert(enc.Encode(data))
	}
}
