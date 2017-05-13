package main

import (
	"net/url"
)

type epArg struct {
	Name  string `json:"name"`
	Usage string `json:"usage"`
}

type endPoint struct {
	EndPoint string  `json:"endpoint"`
	Purpose  string  `json:"purpose"`
	Args     []epArg `json:"args"`
}

var eps []endPoint = []endPoint{
	endPoint{
		EndPoint: "/version",
		Purpose:  "show version info",
		Args:     []epArg{},
	},
}

func help(url.Values) interface{} {
	return eps
}
