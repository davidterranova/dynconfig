package adaptors

import (
	"context"
	"fmt"
	"log"

	"github.com/davidterranova/dynconfig"
	"github.com/fsnotify/fsnotify"
)

type FileWatcherAdaptor struct {
	filePath string
	operator *dynconfig.Operator
	dynconfig.ConfigReader
}

func NewFileWatcherAdaptor(path string, reader dynconfig.ConfigReader) *FileWatcherAdaptor {
	return &FileWatcherAdaptor{
		filePath:     path,
		ConfigReader: reader,
	}
}

func (a *FileWatcherAdaptor) Register(operator *dynconfig.Operator) {
	a.operator = operator
}

func (a FileWatcherAdaptor) Watch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}
	defer watcher.Close()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					a.operator.ConfigChanged(a)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("file watcher error: %s", err)
			}
		}
	}()
	err = watcher.Add(a.filePath)
	if err != nil {
		return fmt.Errorf("failed to watch '%s': %w", a.filePath, err)
	}
	return nil
}
