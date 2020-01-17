package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

type file struct {
	originalPath string
	realPath     string
	signal       syscall.Signal
	pid          int
}
type Config struct {
	// Format: /tmp/foo:USR2:1 (file:signal:pid)
	Watch string
	Debug bool
	files []*file
}

func (c *Config) Files() []*file {
	return c.files
}

func (c *Config) ByOriginalPath(f string) *file {
	for _, v := range c.files {
		if v.realPath == f {
			return v
		}
	}
	return nil
}
func (c *Config) ByRealPath(f string) *file {
	for _, v := range c.files {
		if v.realPath == f {
			return v
		}
	}
	return nil
}

func (c *Config) Parse() {
	list := strings.Split(c.Watch, ",")
	for _, v := range list {
		tmp := strings.Split(v, ":")
		if len(tmp) != 3 {
			log.Fatal("invalid config")
		}
		sig := getSignal(tmp[1])
		if sig == 0 {
			log.Fatalf("unknown signal %s in config", tmp[1])
		}

		f, err := filepath.EvalSymlinks(tmp[0])
		if err != nil {
			log.Fatalf("error finding symlink for %s: %s\n", f, err)
		}

		pid, err := strconv.Atoi(tmp[2])
		if err != nil {
			log.Fatalf("error finding pid in config %s: %s\n", v, err)
		}
		c.files = append(c.files, &file{
			originalPath: tmp[0],
			realPath:     f,
			signal:       sig,
			pid:          pid,
		})
	}
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
