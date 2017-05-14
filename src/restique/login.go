package main

import (
	"net/http"
	"net/url"
)

func login(args url.Values) interface{} {
	name := val(args, "name")
	pass := val(args, "pass")
	code := val(args, "code")
	if name == "" {
		return httpError{
			Code: http.StatusBadRequest,
			Mesg: "[name] not provided",
		}
	}
	if pass == "" && code == "" {
		return httpError{
			Code: http.StatusBadRequest,
			Mesg: "at least one of [pass] or [code] must be provided",
		}
	}
	fail := httpError{
		Code: http.StatusUnauthorized,
		Mesg: "user not found or authentication failed",
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
