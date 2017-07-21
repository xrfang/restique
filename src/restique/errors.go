package main

import (
	"fmt"
	"runtime"
	"strings"
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func catch(err *error) {
	if e := recover(); e != nil {
		*err = e.(error)
	}
}

func trace(msg string, args ...interface{}) (logs []string) {
	msg = fmt.Sprintf(msg, args...)
	logs = []string{msg, ""}
	n := 1
	for {
		n++
		pc, file, line, ok := runtime.Caller(n)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		name := f.Name()
		if strings.HasPrefix(name, "runtime.") {
			continue
		}
		logs = append(logs, fmt.Sprintf("(%s:%d) %s", file, line, name))
	}
	return
}
