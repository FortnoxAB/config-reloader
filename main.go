package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/koding/multiconfig"
)

func main() {
	config := &Config{}
	multiconfig.New().MustLoad(config)
	config.Parse()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

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
					updateWatcher(config, watcher, f)
					config.SignalPid(event.Name)
					continue
				}

				if event.Op == fsnotify.Write {
					config.SignalPid(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			case <-quit:
				log.Println("exiting on signal")
				return
			}
		}
	}()

	addWatchers(watcher, config)
	wg.Wait()
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
}
