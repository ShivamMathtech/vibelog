package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/vibelog/vibelog/internal/store"
)

type Watcher struct {
	SessionID string
	RepoPath  string
	store     *store.Store
	watcher   *fsnotify.Watcher
	stop      chan bool
	events    chan fsnotify.Event
}

func NewWatcher(sessionID, repoPath string, s *store.Store) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		SessionID: sessionID,
		RepoPath:  repoPath,
		store:     s,
		watcher:   watcher,
		stop:      make(chan bool),
		events:    make(chan fsnotify.Event, 100),
	}

	// Watch all subdirectories, skip .git and node_modules
	if err := w.addRecursive(repoPath); err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watcher) addRecursive(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base == ".git" || base == "node_modules" || base == ".vibelog" || base == "vendor" || base == "target" || strings.HasPrefix(base, ".") {
			return filepath.SkipDir
		}
		return w.watcher.Add(path)
	})
}

func (w *Watcher) Start() {
	go w.processEvents()
	go w.watch()
}

func (w *Watcher) watch() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				if !strings.Contains(event.Name, ".git") && !strings.Contains(event.Name, "node_modules") {
					w.events <- event
				}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "watcher error: %v\n", err)
		case <-w.stop:
			return
		}
	}
}

func (w *Watcher) processEvents() {
	timer := time.NewTimer(2 * time.Second)
	timer.Stop()

	var pending []fsnotify.Event

	for {
		select {
		case event := <-w.events:
			pending = append(pending, event)
			timer.Reset(2 * time.Second)
		case <-timer.C:
			if len(pending) > 0 {
				w.flush(pending)
				pending = nil
			}
		case <-w.stop:
			if len(pending) > 0 {
				w.flush(pending)
			}
			return
		}
	}
}

func (w *Watcher) flush(events []fsnotify.Event) {
	var files []string
	seen := make(map[string]bool)
	for _, e := range events {
		if !seen[e.Name] {
			seen[e.Name] = true
			files = append(files, e.Name)
		}
	}

	metadata := fmt.Sprintf(`{"files": [%s], "count": %d}`, joinQuoted(files), len(files))
	content := fmt.Sprintf("File changes detected: %s", strings.Join(files, ", "))

	w.store.AddEvent(w.SessionID, "file_change", content, metadata)
}

func (w *Watcher) Stop() {
	close(w.stop)
	w.watcher.Close()
}

func joinQuoted(files []string) string {
	var parts []string
	for _, f := range files {
		parts = append(parts, fmt.Sprintf(`"%s"`, f))
	}
	return strings.Join(parts, ", ")
}
