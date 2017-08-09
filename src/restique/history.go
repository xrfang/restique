package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type CacheEntry struct {
	Ident    string
	UseCount int
	LastUse  time.Time
	SQL      string
	rawsql   string
}

type MfuCache struct {
	entries map[string]map[string]CacheEntry
	sync.RWMutex
}

func (c *MfuCache) Initialize() {
	c.entries = make(map[string]map[string]CacheEntry)
}

func (c MfuCache) Get(user, sql string) CacheEntry {
	fmt.Println("history:", user)
	sql = strings.TrimSpace(sql)
	rawsql := sql
	ident := ""
	cnt := 1
	if len(sql) > 0 && sql[0] == '#' {
		ss := strings.SplitN(sql, "\n", 2)
		if len(ss) == 2 {
			ident = strings.TrimSpace(ss[0][1:])
			rawsql = strings.TrimSpace(ss[1])
			cnt = 10 //named SQL are 10x more important than unamed ones :-)
		}
	}
	if ident == "" {
		s := rx.ReplaceAllString(rawsql, " ")
		ident = fmt.Sprintf("%x", md5.Sum([]byte(s)))
	}
	var entry CacheEntry
	c.RLock()
	h, ok := c.entries[user]
	c.RUnlock()
	if ok {
		entry, ok = h[ident]
		if ok {
			cnt = 1
		}
	} else {
		c.Lock()
		c.entries[user] = make(map[string]CacheEntry)
		c.Unlock()
	}
	entry.Ident = ident
	entry.rawsql = rawsql
	entry.SQL = sql
	if time.Now().Day() != entry.LastUse.Day() {
		entry.UseCount += cnt
	}
	entry.LastUse = time.Now()
	if entry.rawsql != "" {
		c.Lock()
		c.entries[user][ident] = entry
		c.Unlock()
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "    ")
	e.Encode(mfu.entries)
	return entry
}

var (
	mfu MfuCache
	rx  *regexp.Regexp
)

func init() {
	mfu.Initialize()
	rx = regexp.MustCompile(`(?s)\s+`)
}
