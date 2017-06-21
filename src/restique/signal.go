package main

import (
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func handleSignals() {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGHUP)
	go func() {
		for {
			switch <-sigch {
			case syscall.SIGHUP:
				LoadAuthDb()
				LoadDSNs()
			}
		}
	}()
}

func reloadConfig() {
	hc := http.Client{Timeout: 3 * time.Second}
	resp, err := hc.Get("http://127.0.0.1:" + rc.SERVICE_PORT + "/pid")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 6)
	n, _ := resp.Body.Read(buf)
	if n == 0 {
		return
	}
	pid, _ := strconv.Atoi(string(buf[:n]))
	if pid == 0 {
		return
	}
	p, err := os.FindProcess(pid)
	if err == nil {
		p.Signal(syscall.SIGHUP)
	}
}
