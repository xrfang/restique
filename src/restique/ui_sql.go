package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	QRY_RESULT = `
<div style="position:relative;margin-top:60px;border:dotted lightgray">
<div style="background:{{HINTBG}};padding:6px">{{SUMMARY}}</div>
<pre style="margin:0px;overflow:auto">{{DATA}}</pre>
</div>
`
	QRY_CONTENT = `
<form method="POST" action="/uisql" onsubmit="doQuery()">
<textarea name="sql" id="sql" rows=2 style="display:block;width:100%"
onkeyup="resize(this)" onfocus="resize(this)">{{SQL}}</textarea>
<div style="position:absolute;width:100%">
<span style="float:left">
{{USE}}<input id="qry" style="padding-top:6px;padding-bottom:6px;padding-left:15px;padding-right:15px;margin:10px" type="submit" name="SUBMIT"/>
</span><button id="rx" onclick="toggleHistory()"
type="button" style="padding-top:4px;padding-bottom:4px;padding-left:8px;padding-right:8px;margin-top:10px;font-size:1.1em">&rx;</button>
<span style="float:right;margin-top:10px;margin-right:16px">mode:
<select name="act" style="padding-top:6px;padding-bottom:6px;padding-left:15px;padding-right:15px">
<option {{MODQRY}}>QUERY</option>
<option {{MODEXE}}>EXEC</option>
</select></span><span style="float:right;margin:10px">max height:
<select name="maxh" id="maxh" style="padding:6px">
<option value="12" {{XHS}}>SMALL</option>
<option value="23" {{XHL}}>LARGE</option>
<option value="9999" {{XHU}}>UNLIMITED</option>
</select></span><span style="float:right;margin:10px">result:
<select name="ret" style="padding:6px">
<option value="view" {{XHS}}>view in browser</option>
<option value="csv" {{XHL}}>download CSV</option>
<option value="json" {{XHU}}>download JSON</option>
</select></span>
</div>
</form>
<div style="position:relative;margin-top:60px;margin-bottom:-45px;border:inset
1px;display:none" id="history">
<div class="oddhist" onclick="use(this)">SELECT * FROM table1</div>
<div class="evenhist" onclick="use(this)">SELECT id,updated FROM table2
    WHERE shop_id=102345</div>
</div>
{{RESULT}}
<script>
function doQuery() {
    document.getElementById("qry").style.background = "pink";
}
function resize(a) {
    var rows = a.value.split("\n").length + 1
	var mh = document.getElementById("maxh").value;
	if (rows > mh) rows = mh;
    a.rows = rows
}
function toggleHistory() {
	var pressed = "pink"
    var rx = document.getElementById("rx");
	var hist = document.getElementById("history");
	if (rx.style.background == pressed) {
		rx.style.background = ""
		hist.style.display = "none"
	} else {
		rx.style.background = pressed
		hist.style.display = ""
	}
}
function use(item) {
	var sql = document.getElementById("sql")
	sql.value = item.textContent
	resize(sql)
	toggleHistory()
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
	maxh := args.Get("maxh")
	if maxh == "" {
		c, err := r.Cookie("maxh")
		if err == nil {
			maxh = c.Value
		}
	}
	ret := args.Get("ret")
	act := args.Get("act")
	sql := strings.TrimSpace(args.Get("sql"))

	var (
		qry_res string
		rawdata queryResults
	)
	if args.Get("SUBMIT") != "" {
		http.SetCookie(w, &http.Cookie{
			Name:    "maxh",
			Value:   maxh,
			Path:    "/",
			Expires: time.Now().Add(365 * 24 * time.Hour),
		})
		code := http.StatusOK
		data := ""
		sample := ""
		summary := ""
		hintbg := "lightyellow"
		if sql == "" {
			summary = "empty statement"
			hintbg = "pink"
		} else {
			var res interface{}
			arg := url.Values{
				"use": []string{db},
				"sql": []string{sql},
			}
			if act == "EXEC" {
				res = exec(arg)
			} else {
				res = query(arg)
			}
			if arg.Get("RESTIQUE_MAXROW") == "1" {
				hintbg = "pink"
			}
			switch res.(type) {
			case httpError:
				summary = res.(httpError).Mesg
				hintbg = "pink"
			case queryResults:
				rawdata = res.(queryResults)
				summary = arg.Get("RESTIQUE_SUMMARY")
				http.SetCookie(w, &http.Cookie{
					Name:    "use",
					Value:   db,
					Path:    "/",
					Expires: time.Now().Add(365 * 24 * time.Hour),
				})
				data, sample = tabulate(res.(queryResults), sql)
			default:
				summary = fmt.Sprintf("invalid result type: %T", res)
				hintbg = "pink"
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
				Reply:    sample,
			}
		}
		qry_res = strings.Replace(QRY_RESULT, "{{SUMMARY}}", summary, 1)
		qry_res = strings.Replace(qry_res, "{{DATA}}", data, 1)
		qry_res = strings.Replace(qry_res, "{{HINTBG}}", hintbg, 1)
	}
	use := `<select name="use" style="padding:6px">`
	var dss []string
	for ds := range dsns {
		dss = append(dss, ds)
	}
	sort.Strings(dss)
	for _, ds := range dss {
		if ds == db {
			use += fmt.Sprintf("\n\t"+`<option value="%s" selected>%s</option>`, ds, ds)
		} else {
			use += fmt.Sprintf("\n\t"+`<option value="%s">%s</option>`, ds, ds)
		}
	}
	switch ret {
	case "csv":
		fn := fmt.Sprintf("restique_query_result_%s.csv",
			time.Now().Format("2006-01-02_15.04.05"))
		w.Header().Add("Content-Disposition", "attachment; filename="+fn)
		if len(rawdata) > 0 {
			enc := csv.NewWriter(w)
			enc.UseCRLF = true
			var cols []string
			keys := getKeys(rawdata, sql)
			for _, k := range keys {
				cols = append(cols, k.key)
			}
			assert(enc.Write(cols))
			for _, rd := range rawdata {
				var row []string
				for _, c := range cols {
					row = append(row, fmt.Sprintf("%v", rd[c]))
				}
				assert(enc.Write(row))
			}
			enc.Flush()
		}
	case "json":
		fn := fmt.Sprintf("restique_query_result_%s.json",
			time.Now().Format("2006-01-02_15.04.05"))
		w.Header().Add("Content-Disposition", "attachment; filename="+fn)
		enc := json.NewEncoder(w)
		enc.SetIndent("", "    ")
		assert(enc.Encode(rawdata))
	default:
		modqry := "selected"
		modexe := ""
		if act == "EXEC" {
			modqry, modexe = modexe, modqry
		}
		var xh_small, xh_large, xh_nolimit string
		switch maxh {
		case "12":
			xh_small = "selected"
		case "23":
			xh_large = "selected"
		default:
			xh_nolimit = "selected"
		}
		body := strings.Replace(QRY_CONTENT, "{{USE}}", use, 1)
		body = strings.Replace(body, "{{SQL}}", sql, 1)
		body = strings.Replace(body, "{{MODQRY}}", modqry, 1)
		body = strings.Replace(body, "{{MODEXE}}", modexe, 1)
		body = strings.Replace(body, "{{XHS}}", xh_small, 1)
		body = strings.Replace(body, "{{XHL}}", xh_large, 1)
		body = strings.Replace(body, "{{XHU}}", xh_nolimit, 1)
		page := strings.Replace(PAGE, "{{VERSION}}", fmt.Sprintf("V%s.%s",
			_G_REVS, _G_HASH), 1)
		page = strings.Replace(page, "{{CONTENT}}", body, 1)
		page = strings.Replace(page, "{{RESULT}}", qry_res, 1)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page)
	}
}
