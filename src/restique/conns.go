package main

import (
	"net/url"
)

func conns(url.Values) interface{} {
	var cs []map[string]string
	for k, v := range dsns {
		cs = append(cs, map[string]string{
			"name":        k,
			"driver":      v.Driver,
			"description": v.Memo,
		})
	}
	return cs
}
