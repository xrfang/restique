package main

import (
	"fmt"
	"net/http"
	"strings"
)

const LGN_CONTENT = `
<div class="form" style="top:20%">
  <form method=POST action="/loginui">
    <input type="text" name="name" placeholder="username" value="{{name}}"/>
    <input type="password" name="code" placeholder="OTP code"/>
    <input type="password" name="pass" placeholder="password"/>
    <div style="padding-bottom:10px;color:red">{{err}}</div>
    <button>login</button>
  </form>
</div>
`

func uiLgn(w http.ResponseWriter, r *http.Request) {
	if sessions.Validate(r) {
		http.Redirect(w, r, "/uisql", http.StatusSeeOther)
		return
	}
	page := strings.Replace(PAGE, "{{VERSION}}", fmt.Sprintf("V%s.%s",
		_G_REVS, _G_HASH), 1)
	html := strings.Replace(page, "{{CONTENT}}", LGN_CONTENT, 1)
	html = strings.Replace(html, "{{name}}", r.URL.Query().Get("name"), 1)
	html = strings.Replace(html, "{{err}}", r.URL.Query().Get("err"), 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}
