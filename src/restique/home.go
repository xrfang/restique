package main

import (
	"net/http"
	"net/url"
	"strings"
)

func home(args url.Values) interface{} {
	path := val(args, "REQUEST_URL_PATH")
	if path == "/" {
		panic(httpError{Code: http.StatusSeeOther, Mesg: "/uilgn"})
	}
	switch {
	case strings.HasPrefix(path, "/query/"):
		ua := strings.SplitN(path[7:], "/", 2)
		switch len(ua) {
		case 1:
			args["use"] = []string{ua[0]}
		case 2:
			args["use"] = []string{ua[0]}
			args["sql"] = []string{ua[1]}
		}
		return query(args)
	case strings.HasPrefix(path, "/exec/"):
		ua := strings.SplitN(path[6:], "/", 2)
		switch len(ua) {
		case 1:
			args["use"] = []string{ua[0]}
		case 2:
			args["use"] = []string{ua[0]}
			args["sql"] = []string{ua[1]}
		}
		return exec(args)
	}
	return httpError{
		Code: http.StatusNotFound,
		Mesg: "not found: " + path,
	}
}
