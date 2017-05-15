package main

import (
	"net"
	"net/http"
	"strings"
)

var allowed_cidrs []string

func AccessDenied(r *http.Request) bool {
	if len(allowed_cidrs) == 0 {
		return false
	}
	ip := strings.Split(r.RemoteAddr, ":")[0]
	addr := net.ParseIP(ip)
	for _, cli := range allowed_cidrs {
		_, cidr, err := net.ParseCIDR(cli)
		assert(err)
		if cidr.Contains(addr) {
			return false
		}
	}
	return true
}
