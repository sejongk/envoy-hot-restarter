package watcher

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

type FileSystemEventHandler interface {
	On_created() error
	On_modified() error
	On_deleted() error
	On_moved() error
}

func Watch(path string, handler FileSystemEventHandler) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					err = handler.On_modified()
				} else if event.Op&fsnotify.Create == fsnotify.Create {
					err = handler.On_created()
				} else if event.Op&fsnotify.Remove == fsnotify.Remove {
					err = handler.On_deleted()
				} else if event.Op&fsnotify.Rename == fsnotify.Rename {
					err = handler.On_moved()
				}
				log.Error(err)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error(err)
			}

		}
	}()
	err = watcher.Add(path)
	if err != nil {
		log.Error(err)
	}
	<-done
}
