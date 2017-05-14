package main

import (
	"encoding/json"
	"os"

	"github.com/mdp/qrterminal"
	"github.com/pquerna/otp/totp"
)

type authInfo struct {
	Name   string `json:"name"`
	Pass   string `json:"pass"`
	Secret string `json:"secret"`
}

func (ai authInfo) Validate(code string) bool {
	return totp.Validate(code, ai.Secret)
}

var authDb map[string]authInfo

func init() {
	authDb = make(map[string]authInfo)
}

func SetAuth(user, pass string) {
	gopts := totp.GenerateOpts{Issuer: self, AccountName: user}
	key, err := totp.Generate(gopts)
	assert(err)
	ai := authDb[user]
	ai.Name = user
	ai.Pass = pass
	ai.Secret = key.Secret()
	authDb[user] = ai
	f, err := os.Create(rc.AUTH_PATH)
	assert(err)
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	assert(enc.Encode(authDb))
	qrterminal.Generate(key.String(), qrterminal.L, os.Stdout)
	return
}

func LoadAuthDb() {
	f, err := os.Open(rc.AUTH_PATH)
	assert(err)
	defer f.Close()
	dec := json.NewDecoder(f)
	assert(dec.Decode(&authDb))
}
