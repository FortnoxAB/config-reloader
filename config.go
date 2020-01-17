package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type Config struct {
	// Format: /tmp/foo:USR2:1 (file:signal:pid)
	Watch string
	Debug bool
}

func (c *Config) Files() []string {
	list := strings.Split(c.Watch, ",")
	files := []string{}
	for _, v := range list {
		tmp := strings.Split(v, ":")
		if len(tmp) != 3 {
			log.Fatal("invalid config")
		}

		if sig := getSignal(tmp[1]); sig == 0 {
			log.Fatalf("unknown signal %s in config", tmp[1])
		}

		files = append(files, tmp[0])
	}
	return files
}

func (c *Config) FileExists(file string) bool {
	for _, v := range c.Files() {
		if v == file {
			return true
		}
	}
	return false
}
func (c *Config) SignalPid(file string) {
	list := strings.Split(c.Watch, ",")
	var v string
	for _, v = range list {
		if strings.HasPrefix(v, file) {
			break
		}
	}
	tmp := strings.Split(v, ":")
	pid, err := strconv.Atoi(tmp[2])
	signal := tmp[1]
	if err != nil {
		log.Printf("error finding pid in config %s: %s\n", v, err)
		return
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		log.Printf("error finding pid %d: %s\n", pid, err)
		return
	}

	err = p.Signal(getSignal(signal))
	if err != nil {
		log.Printf("error sending signal to pid %d: %s\n", pid, err)
		return
	}
	log.Printf("sent signal %s to %d\n", signal, pid)
}

func getSignal(s string) syscall.Signal {
	switch s {
	case "USR1":
		return syscall.SIGUSR1
	case "USR2":
		return syscall.SIGUSR2
	case "INT":
		return syscall.SIGINT
	case "KILL":
		return syscall.SIGKILL
	}

	return 0
}
