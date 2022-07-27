package filewatcher

import (
	"fmt"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher *fsnotify.Watcher
}

func New(filePath string) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("unable to create new file watcher: %w", err)
	}

	if err := watcher.Add(filePath); err != nil {
		return nil, fmt.Errorf("unable to add file %s to watcher: %w", filePath, err)
	}

	return &FileWatcher{
		watcher: watcher,
	}, nil
}

func (f *FileWatcher) Watch(fu func(), cancelCh chan os.Signal) {
	for {
		select {
		case event, ok := <-f.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				fu()
			}
		case err, ok := <-f.watcher.Errors:
			if !ok {
				// channel closed
				return
			}
			if err != nil {
				log.Println("error:", err)
			}
		case <-cancelCh:
			f.Close()
			return
		}
	}
}

func (f *FileWatcher) Close() {
	f.watcher.Close()
}
