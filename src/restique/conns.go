package main

import "net/url"

func conns(url.Values) interface{} {
	var cs []map[string]string
	for k, v := range dsns {
		info := map[string]string{
			"name":        k,
			"driver":      v.Driver,
			"description": v.Memo,
		}
		if v.Driver == "[multi]" {
			info["dsn"] = v.Dsn
		}
		cs = append(cs, info)
	}
	return cs
}
