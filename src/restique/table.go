package main

import (
	"fmt"
	"html"
	"sort"
	"strings"
)

type (
	queryResult  map[string]interface{}
	queryResults []queryResult
)

func tabulate(data queryResults, query string) (string, string) {
	if len(data) == 0 {
		return "", ""
	}
	type keyInfo struct {
		key string
		pos int
	}
	var keys []keyInfo
	for k := range data[0] {
		keys = append(keys, keyInfo{key: k, pos: strings.Index(query, k)})
	}
	sort.Slice(keys, func(i, j int) bool {
		pi := keys[i].pos
		pj := keys[j].pos
		if pi < 0 {
			if pj >= 0 {
				return false //key found in query always returns first
			}
			//otherwise, ordering by key alphabetically
			return keys[i].key < keys[j].key
		} else if pj < 0 {
			return true //key found in query always return first
		}
		if pi == pj {
			return keys[i].key < keys[j].key
		}
		return pi < pj
	})
	dat := []string{}
	sample := func(vals []string) {
		if len(dat) > 6 {
			return
		}
		str := strings.Join(vals, ",")
		if len(str) > 100 {
			str = str[:100] + "..."
		}
		dat = append(dat, str)
	}
	tab := []string{`<table border="0" width="100%"><tr class="headrow">`}
	fields := []string{}
	for _, k := range keys {
		fields = append(fields, k.key)
		tab = append(tab, `<th class="thcell">`+k.key+`</th>`)
	}
	sample(fields)
	tab = append(tab, `</tr>`)
	for i, d := range data {
		if i%2 == 0 {
			tab = append(tab, `<tr class="evenrow">`)
		} else {
			tab = append(tab, `<tr class="oddrow">`)
		}
		fields = []string{}
		for _, k := range keys {
			v := fmt.Sprintf("%v", d[k.key])
			fields = append(fields, v)
			v = html.EscapeString(v)
			tab = append(tab, `<td class="tdcell"><pre>`+v+`</pre></td>`)
		}
		sample(fields)
		tab = append(tab, `</tr>`)
	}
	if len(dat) > 5 {
		dat[4] = "... ..."
		dat = dat[:5]
	}
	return strings.Join(tab, "\n"), strings.Join(dat, "\n")
}
