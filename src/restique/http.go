package main

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func handler(proc func(url.Values) interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				//TODO: logging?
				http.Error(w, e.(error).Error(), http.StatusInternalServerError)
			}
		}()
		//TODO: access control
		var args url.Values
		if r.Method == "POST" || r.Method == "PUT" {
			r.ParseForm()
			args = r.Form
		} else {
			args = r.URL.Query()
		}
		data := proc(args)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		assert(enc.Encode(data))
	}
}
