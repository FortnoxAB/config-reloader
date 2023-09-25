package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

type file struct {
	originalPath string
	realPath     string
	signal       syscall.Signal
	pidOrPath    string
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
		sig := unix.SignalNum("SIG" + tmp[1])
		if sig == 0 {
			log.Fatalf("unknown signal %s in config", tmp[1])
		}

		f, err := filepath.EvalSymlinks(tmp[0])
		if err != nil {
			log.Fatalf("error finding symlink for %s: %s\n", f, err)
		}

		c.files = append(c.files, &file{
			originalPath: tmp[0],
			realPath:     f,
			signal:       sig,
			pidOrPath:    tmp[2],
		})
	}
}

func (c *Config) SignalPid(path string) {
	var v *file
	for _, v = range c.Files() {
		if v.realPath == path {
			break
		}
	}

	pid, err := getPID(v.pidOrPath)
	if err != nil {
		log.Println(err)
		return
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		log.Printf("error finding pid %d: %s\n", pid, err)
		return
	}

	err = p.Signal(v.signal)
	if err != nil {
		log.Printf("error sending signal to pid %d: %s\n", pid, err)
		return
	}
	log.Printf("sent signal %s to %d\n", v.signal, pid)
}

// getPID returns PID as integer from string or from a pidfile
func getPID(s string) (int, error) {
	pid, err := strconv.Atoi(s)
	if err != nil {
		//try to find using pid file instead
		content, err := os.ReadFile(s)
		if err != nil {
			return 0, fmt.Errorf("unable to open pidfile %s: %w", s, err)
		}
		pid, err = strconv.Atoi(strings.TrimSpace(string(content)))
		if err != nil {
			return 0, fmt.Errorf("error finding pid in config %s: %w", s, err)
		}
	}
	return pid, nil
}
