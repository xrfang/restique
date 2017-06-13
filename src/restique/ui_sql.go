package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const CONTENT = `
<form method="POST" action="/uisql">
<input type="hidden" name="action" value="{{ACTION}}"/>
<textarea style="display:block;width:100%%;height:80%" name="sql" id="sql" rows=5 onkeyup="resizeTextarea('sql')"></textarea>
{{USE}}<input style="padding-top:6px;padding-bottom:6px;padding-left:15px;padding-right:15px;margin:10px" type="submit" name="SUBMIT"/>
</form>
<script>
function resizeTextarea (id) {
  var a = document.getElementById(id);
  a.style.height = 'auto';
  a.style.height = (a.scrollHeight+10)+'px';
}
</script>
`

func uiSql(w http.ResponseWriter, r *http.Request) {
	if AccessDenied(r) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}
	if !sessions.SessionOK(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	requestTime := time.Now()
	args := r.URL.Query()
	if r.Method == "POST" || r.Method == "PUT" {
		r.ParseForm()
		args = r.Form
	}
	db := args.Get("use")
	act := args.Get("action")
	sql := args.Get("sql")
	if args.Get("SUBMIT") != "" {
		var (
			res interface{}
			out bytes.Buffer
		)
		arg := map[string][]string{
			"use": []string{db},
			"sql": []string{sql},
		}
		if act == "exec" {
			res = exec(arg)
		} else {
			res = query(arg)
		}
		code := http.StatusOK
		data := ""
		switch res.(type) {
		case httpError:
			code = res.(httpError).Code
			data = res.(httpError).Mesg
			http.Error(w, data, code)
		default:
			mw := io.MultiWriter(&out, w)
			enc := json.NewEncoder(mw)
			enc.SetIndent("", "    ")
			err := enc.Encode(res)
			if err != nil {
				code = http.StatusInternalServerError
				data = err.Error()
				http.Error(w, data, code)
			} else {
				data = out.String()
			}
		}
		delete(args, "REQUEST_URL_PATH")
		lms <- logMessage{
			Client:   r.RemoteAddr,
			Time:     requestTime,
			Duration: time.Since(requestTime).Seconds(),
			Request:  r.URL.Path,
			Params:   args,
			Cookie:   r.Cookies(),
			Code:     code,
			Reply:    data,
		}
		return
	}
	use := `<select name="use" style="padding:6px">`
	for ds := range dsns {
		if ds == db {
			use += fmt.Sprintf("\n\t"+`<option value="%s" selected="selected">%s</option>`, ds, ds)
		} else {
			use += fmt.Sprintf("\n\t"+`<option value="%s">%s</option>`, ds, ds)
		}
	}
	body := strings.Replace(CONTENT, "{{USE}}", use, 1)
	body = strings.Replace(body, "{{ACTION}}", act, 1)
	html := strings.Replace(PAGE, "{{CONTENT}}", body, 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, html)
}
