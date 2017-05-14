package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

var (
	_G_HASH string
	_G_REVS string
)

func version(url.Values) interface{} {
	self := filepath.Base(os.Args[0])
	return map[string]string{
		"app": fmt.Sprintf("%s - RESTful MySQL query proxy", self),
		"ver": fmt.Sprintf("V%s.%s", _G_REVS, _G_HASH),
	}
}
