package main

import (
	"net/http"
	"net/url"
)

func login(args url.Values) interface{} {
	name := args.Get("name")
	pass := args.Get("pass")
	code := args.Get("code")
	if name == "" {
		return httpError{
			Code: http.StatusBadRequest,
			Mesg: "username not provided",
		}
	}
	if pass == "" && code == "" {
		return httpError{
			Code: http.StatusBadRequest,
			Mesg: "password or OTP code required",
		}
	}
	fail := httpError{
		Code: http.StatusUnauthorized,
		Mesg: "authentication failed",
	}
	ai, ok := authDb[name]
	if !ok {
		return fail
	}
	if !ai.Validate(pass, code) {
		return fail
	}
	return nil
}
