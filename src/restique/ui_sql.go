package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	QRY_RESULT = `
<div style="position:relative;margin-top:60px;border:dotted lightgray">
<div style="background:lightyellow;padding:6px">{{SUMMARY}}</div>
<pre style="padding:6px;overflow:auto">{{DATA}}</pre>
</div>
`
	QRY_CONTENT = `
<form method="POST" action="/uisql">
<textarea style="display:block;width:100%%" name="sql" id="sql" rows=5 onkeyup="resize('sql')">{{SQL}}</textarea>
<div style="position:absolute;width:100%%">
<span style="float:left">
{{USE}}<input style="padding-top:6px;padding-bottom:6px;padding-left:15px;padding-right:15px;margin:10px" type="submit" name="SUBMIT"/>
</span>
<span style="float:right;margin-top:10px;margin-right:16px">mode:
<select name="act" style="padding-top:6px;padding-bottom:6px;padding-left:15px;padding-right:15px">
<option {{MODQRY}}>QUERY</option>
<option {{MODEXE}}>EXEC</option>
</select>
</span>
</div>
</form>
{{RESULT}}
<script>
function resize(id) {
  var a = document.getElementById(id);
  a.style.height = 'auto';
  a.style.height = (a.scrollHeight+10)+'px';
}
</script>
`
)

func uiSql(w http.ResponseWriter, r *http.Request) {
	if AccessDenied(r) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}
	if !sessions.SessionOK(r) {
		http.Redirect(w, r, "/uilgn", http.StatusSeeOther)
		return
	}
	requestTime := time.Now()
	args := r.URL.Query()
	if r.Method == "POST" || r.Method == "PUT" {
		r.ParseForm()
		args = r.Form
	}
	db := args.Get("use")
	if db == "" {
		c, err := r.Cookie("use")
		if err == nil {
			db = c.Value
		}
	}
	act := args.Get("act")
	sql := args.Get("sql")
	var qry_res string
	if args.Get("SUBMIT") != "" {
		var (
			res interface{}
			out bytes.Buffer
		)
		arg := url.Values{
			"use": []string{db},
			"sql": []string{sql},
		}
		if act == "EXEC" {
			res = exec(arg)
		} else {
			res = query(arg)
		}
		code := http.StatusOK
		data := ""
		summary := ""
		switch res.(type) {
		case httpError:
			summary = res.(httpError).Mesg
		default:
			summary = arg.Get("RESTIQUE_SUMMARY")
			http.SetCookie(w, &http.Cookie{
				Name:    "use",
				Value:   db,
				Path:    "/",
				Expires: time.Now().Add(365 * 24 * time.Hour),
			})
			enc := json.NewEncoder(&out)
			enc.SetIndent("", "    ")
			err := enc.Encode(res)
			if err != nil {
				summary = err.Error()
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
		qry_res = strings.Replace(QRY_RESULT, "{{SUMMARY}}", summary, 1)
		qry_res = strings.Replace(qry_res, "{{DATA}}", data, 1)
	}
	use := `<select name="use" style="padding:6px">`
	for ds := range dsns {
		if ds == db {
			use += fmt.Sprintf("\n\t"+`<option value="%s" selected>%s</option>`, ds, ds)
		} else {
			use += fmt.Sprintf("\n\t"+`<option value="%s">%s</option>`, ds, ds)
		}
	}
	modqry := "selected"
	modexe := ""
	if act == "EXEC" {
		modqry, modexe = modexe, modqry
	}
	body := strings.Replace(QRY_CONTENT, "{{USE}}", use, 1)
	body = strings.Replace(body, "{{SQL}}", sql, 1)
	body = strings.Replace(body, "{{MODQRY}}", modqry, 1)
	body = strings.Replace(body, "{{MODEXE}}", modexe, 1)
	page := strings.Replace(PAGE, "{{VERSION}}", fmt.Sprintf("V%s.%s",
		_G_REVS, _G_HASH), 1)
	page = strings.Replace(page, "{{CONTENT}}", body, 1)
	page = strings.Replace(page, "{{RESULT}}", qry_res, 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, page)
}
