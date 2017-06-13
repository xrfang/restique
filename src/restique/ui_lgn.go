package main

import (
	"fmt"
	"net/http"
	"strings"
)

const LGN_CONTENT = `
<div class="form" style="top:20%%">
  <form method=POST action="/loginui">
    <input type="text" name="name" placeholder="username"/>
	<input type="password" name="code" placeholder="OTP code"/>
    <input type="password" name="pass" placeholder="password"/>
    <button>login</button>
  </form>
</div>
`

func uiLgn(w http.ResponseWriter, r *http.Request) {
	html := strings.Replace(PAGE, "{{CONTENT}}", LGN_CONTENT, 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, html)
}
