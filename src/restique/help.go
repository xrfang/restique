package main

import (
	"net/http"
	"net/url"
	"strings"
)

type epArg struct {
	Name string `json:"name"`
	Hint string `json:"hint"`
}

type endPoint struct {
	EndPoint string  `json:"endpoint"`
	Purpose  string  `json:"purpose"`
	Note     string  `json:"note,omitempty"`
	Args     []epArg `json:"args,omitempty"`
	Returns  []epArg `json:"returns,omitempty"`
}

var eps []endPoint = []endPoint{
	endPoint{
		EndPoint: "/conns",
		Purpose:  "list available database connections",
	},
	endPoint{
		EndPoint: "/exec",
		Purpose:  "execute SQL statement",
		Note:     "alternative form: /exec/<use>/<sql>",
		Args: []epArg{
			epArg{
				Name: "use",
				Hint: "name of connection to use",
			},
			epArg{
				Name: "sql",
				Hint: "SQL statement (insert/update etc.)",
			},
		},
	},
	endPoint{
		EndPoint: "/login",
		Purpose:  "user authentication",
		Args: []epArg{
			epArg{
				Name: "name",
				Hint: "username",
			},
			epArg{
				Name: "pass",
				Hint: "password",
			},
			epArg{
				Name: "code",
				Hint: "OTP access code",
			},
		},
		Returns: []epArg{
			epArg{
				Name: "session",
				Hint: "session ID (will also be sent via cookie)",
			},
		},
	},
	endPoint{
		EndPoint: "/query",
		Purpose:  "execute SQL query",
		Note:     "alternative form: /query/<use>/<sql>",
		Args: []epArg{
			epArg{
				Name: "use",
				Hint: "name of connection to use",
			},
			epArg{
				Name: "sql",
				Hint: "SQL statement (query only)",
			},
		},
	},
	endPoint{
		EndPoint: "/version",
		Purpose:  "show version info",
	},
}

func help(args url.Values) interface{} {
	path := val(args, "REQUEST_URL_PATH")
	if path == "/" {
		return eps
	}
	switch {
	case strings.HasPrefix(path, "/query/"):
		args := strings.SplitN(path[7:], "/", 2)
		if len(args) != 2 {
			return httpError{
				Code: http.StatusBadRequest,
				Mesg: "missing [use] or [sql]",
			}
		}
		return query(map[string][]string{
			"use": []string{args[0]},
			"sql": []string{args[1]},
		})
	case strings.HasPrefix(path, "/exec/"):
		args := strings.SplitN(path[6:], "/", 2)
		if len(args) != 2 {
			return httpError{
				Code: http.StatusBadRequest,
				Mesg: "missing [use] or [sql]",
			}
		}
		return exec(map[string][]string{
			"use": []string{args[0]},
			"sql": []string{args[1]},
		})
	}
	return httpError{
		Code: http.StatusNotFound,
		Mesg: "not found: " + path,
	}
}
