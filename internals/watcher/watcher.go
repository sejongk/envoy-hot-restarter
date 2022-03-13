package watcher

import (
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileSystemEventHandler interface {
	On_created() error
	On_modified() error
	On_deleted() error
}

func Watch(path string, handler FileSystemEventHandler) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		var (
			timer     *time.Timer
			lastEvent fsnotify.Event
		)
		timer = time.NewTimer(time.Millisecond)
		<-timer.C // timer is expired at first
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				lastEvent = event
				timer.Reset(time.Millisecond * 100)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logger.Error(err)
			case <-timer.C:
				if lastEvent.Op&fsnotify.Write == fsnotify.Write {
					err = handler.On_modified()
				} else if lastEvent.Op&fsnotify.Create == fsnotify.Create {
					err = handler.On_created()
				} else if lastEvent.Op&fsnotify.Remove == fsnotify.Remove {
					err = handler.On_deleted()
					if err != nil {
						logger.Error(err)
					}
					err = watcher.Add(path) // re-register target file to watcher for continuous watching
					if err != nil {
						logger.Error(err)
					}
				}
				if err != nil {
					logger.Error(err)
				}
				// time.Sleep(time.Minute) // sleep 1 min for new envoy to initialize itself
			}

		}
	}()
	err = watcher.Add(path)
	if err != nil {
		logger.Error(err)
	}
	<-done
}
