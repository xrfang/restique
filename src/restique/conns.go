package main

import (
	"net/url"
)

func conns(url.Values) interface{} {
	var cs []map[string]interface{}
	for k, v := range dsns {
		cs = append(cs, map[string]interface{}{
			"name":        k,
			"driver":      v.Driver,
			"description": v.Memo,
			"active":      v.conn != nil,
		})
	}
	return cs
}
