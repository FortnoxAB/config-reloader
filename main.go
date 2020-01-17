package main

import (
	"log"
	"os"
	"os/signal"
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

				/*
					if event.Op == fsnotify.Remove {
						err = watcher.Remove(event.Name)
						if err != nil {
							log.Println(err)
							continue
						}
					}
				*/

				if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
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

func addWatchers(w *fsnotify.Watcher, config *Config) {
	for _, v := range config.Files() {
		log.Printf("adding watcher for %s ", v.realPath)
		err := w.Add(v.realPath)
		if err != nil {
			log.Printf("error adding watcher %s: %s\n", v.realPath, err)
		}
	}
}
