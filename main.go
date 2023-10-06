package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/jonaz/gograce"
	"github.com/koding/multiconfig"
)

var watcherCount int64

func main() {
	config := &Config{}
	multiconfig.New().MustLoad(config)
	config.Parse()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	srv, shutdown := gograce.NewServerWithTimeout(5 * time.Second)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if config.Debug {
					log.Printf("fsnotify: %s", event)
				}

				if event.Op == fsnotify.Remove { // support the way k8s changes mounted configmaps
					f := config.ByRealPath(event.Name)
					if f == nil {
						continue
					}
					atomic.AddInt64(&watcherCount, -1)
					updateWatcher(config, watcher, f)
					config.SignalPid(f.realPath)
					continue
				}

				if event.Op == fsnotify.Write {
					config.SignalPid(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			case <-shutdown:
				log.Println("exiting on signal")
				return
			}
		}
	}()

	addWatchers(watcher, config)

	http.HandleFunc("/health", healthHandler)

	srv.Handler = http.DefaultServeMux
	srv.Addr = ":8080"

	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Println(err)
	}
	wg.Wait()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if watcherCount <= 0 {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "no watchers found!")
	}
	fmt.Fprintf(w, "ok")
}

func updateWatcher(config *Config, w *fsnotify.Watcher, f *file) {
	if config.Debug {
		log.Printf("updateWatcher: %#v\n", f)
	}

	rp, err := filepath.EvalSymlinks(f.originalPath)
	if err != nil {
		log.Println("error updateWatcher find realPath:", err)
		return
	}
	f.realPath = rp
	addWatcher(w, f.realPath)
	if config.Debug {
		log.Printf("added watcher: %#v\n", f)
	}
}

func addWatchers(w *fsnotify.Watcher, config *Config) {
	for _, v := range config.Files() {
		addWatcher(w, v.realPath)
	}
}
func addWatcher(w *fsnotify.Watcher, path string) {
	log.Printf("adding watcher for %s ", path)
	err := w.Add(path)
	if err != nil {
		log.Printf("error adding watcher %s: %s\n", path, err)
	}
	atomic.AddInt64(&watcherCount, 1)
}
