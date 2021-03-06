package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
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

func (c MfuCache) Load(user string) {
	c.Lock()
	defer c.Unlock()
	defer func() {
		if e := recover(); e != nil {
			c.entries[user] = make(map[string]CacheEntry)
		}
	}()
	f, err := os.Open(path.Join(rc.HIST_PATH, user+".json"))
	assert(err)
	defer f.Close()
	dec := json.NewDecoder(f)
	var entries map[string]CacheEntry
	assert(dec.Decode(&entries))
	c.entries[user] = entries
}

func (c MfuCache) Save(user string, entries map[string]CacheEntry) {
	c.Lock()
	c.entries[user] = entries
	c.Unlock()
	f, err := os.Create(path.Join(rc.HIST_PATH, user+".json"))
	if err != nil {
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	enc.Encode(entries)
}

func (c MfuCache) Get(user, sql string) CacheEntry {
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
		c.Load(user)
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
	return entry
}

func (c MfuCache) Update(user string, max int) []CacheEntry {
	c.Lock()
	entries := c.entries[user]
	c.Unlock()
	var res []CacheEntry
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	for _, ce := range entries {
		if ce.LastUse.Before(today) {
			ce.UseCount -= 1
			ce.LastUse = today
		}
		res = append(res, ce)
	}
	sort.Slice(res, func(i, j int) bool {
		udiff := res[i].UseCount - res[j].UseCount
		if udiff > 0 {
			return true
		} else if udiff < 0 {
			return false
		}
		return res[i].LastUse.After(res[j].LastUse)
	})
	if len(res) > max {
		res = res[:max]
	}
	entries = make(map[string]CacheEntry)
	for _, ce := range res {
		entries[ce.Ident] = ce
	}
	c.Save(user, entries)
	return res
}

var (
	mfu MfuCache
	rx  *regexp.Regexp
)

func init() {
	mfu.Initialize()
	rx = regexp.MustCompile(`(?s)\s+`)
}
