package main

import (
	"encoding/json"
	"os"

	"github.com/mdp/qrterminal"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type authInfo struct {
	Name   string `json:"name"`
	Pass   string `json:"pass"`
	Secret string `json:"secret"`
}

func (ai authInfo) Validate(pass, code string) bool {
	if !totp.Validate(code, ai.Secret) {
		return false
	}
	if ai.Pass == "" {
		return true
	}
	return nil == bcrypt.CompareHashAndPassword([]byte(ai.Pass), []byte(pass))
}

var authDb map[string]authInfo

func init() {
	authDb = make(map[string]authInfo)
}

func SetAuth(user, pass string) {
	gopts := totp.GenerateOpts{
		AccountName: user,
		Digits:      otp.Digits(rc.OTP_DIGITS),
		Issuer:      rc.OTP_ISSUER,
		Period:      rc.OTP_TIMEOUT,
	}
	key, err := totp.Generate(gopts)
	assert(err)
	ai := authDb[user]
	ai.Name = user
	ai.Pass = ""
	if pass != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(pass), 12)
		assert(err)
		ai.Pass = string(hash)
	}
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
