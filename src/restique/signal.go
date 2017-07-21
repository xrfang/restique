package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"
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

func savePid() {
	f, err := os.Create(rc.PID_FILE)
	assert(err)
	defer f.Close()
	_, err = f.Write([]byte(strconv.Itoa(os.Getpid())))
	assert(err)
}

func reloadConfig() {
	f, err := os.Open(rc.PID_FILE)
	assert(err)
	defer f.Close()
	buf := make([]byte, 16)
	n, err := f.Read(buf)
	assert(err)
	pid, _ := strconv.Atoi(string(buf[:n]))
	p, err := os.FindProcess(pid)
	assert(err)
	p.Signal(syscall.SIGHUP)
}
