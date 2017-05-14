package main

import (
	"net/url"
)

type epArg struct {
	Name string `json:"name"`
	Hint string `json:"hint"`
}

type endPoint struct {
	EndPoint string  `json:"endpoint"`
	Purpose  string  `json:"purpose"`
	Note     string  `json:"note,omitempty"`
	Args     []epArg `json:"args"`
	Returns  []epArg `json:"returns,omitempty"`
}

var eps []endPoint = []endPoint{
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
		Args: []epArg{
			epArg{
				Name: "use",
				Hint: "database name",
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
		Args:     []epArg{},
	},
}

func help(url.Values) interface{} {
	return eps
}
