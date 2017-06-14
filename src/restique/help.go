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
		EndPoint: "/reload",
		Purpose:  "reload the auth and dsn database",
		Note:     "this API must be called from the localhost",
	},
	endPoint{
		EndPoint: "/version",
		Purpose:  "show version info",
	},
}

func help(args url.Values) interface{} {
	return eps
}
