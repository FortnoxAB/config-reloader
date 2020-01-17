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
				if !config.FileExists(event.Name) {
					continue
				}
				if config.Debug {
					log.Printf("fsnotify: %s", event)
				}
				if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
					config.SignalPid(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Println("error:", err)
				}
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

func addWatchers(w *fsnotify.Watcher, config *Config) {
	for _, v := range config.Files() {
		dir := filepath.Dir(v)
		log.Printf("adding watcher for %s", dir)
		err := w.Add(dir)
		if err != nil {
			log.Printf("error adding watcher %s: %s\n", v, err)
		}
	}
}
