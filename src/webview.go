package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/webview/webview_go"
)

var (
	rotIdx    int
	cWB       webview.WebView
	scrollPos int64
	lock      sync.Mutex
)

func hashedFilePath(path string) string {
	h := sha1.New()
	h.Write([]byte(path))
	return hex.EncodeToString(h.Sum(nil))
}

func writeRotatingHTML(path string, content []byte, rotation int) (string, error) {
	tmpDir := os.TempDir()
	hash := hashedFilePath(path)
	file := filepath.Join(tmpDir, fmt.Sprintf("mdlp-%s-%d.html", hash, rotation))
	err := os.WriteFile(file, content, 0644)
	return file, err
}

func LaunchViewer(path string) {
	html, err := RenderMarkdown(path)
	if err != nil {
		log.Fatalf("Failed to render markdown: %v", err)
	}

	rotPath, err := writeRotatingHTML(path, []byte(html), 0)
	if err != nil {
		log.Fatalf("Failed to write HTML: %v", err)
	}

	w := webview.New(true)
	cWB = w
	defer w.Destroy()

	_ = w.Bind("LoadScroll", func() (int64, error) {
		lock.Lock()
		y := scrollPos
		lock.Unlock()
		return y, nil
	})

	_ = w.Bind("SaveScroll", func(y int64) error {
		lock.Lock()
		scrollPos = y
		lock.Unlock()
		return nil
	})

	w.SetTitle("Markdown Live Preview")
	w.SetSize(800, 600, webview.HintNone)
	w.SetSize(200, 200, webview.HintMin)
	w.Navigate("file://" + rotPath)

	setAppIcon(w)

	go watchAndReload(path)

	w.Run()
}

func watchAndReload(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer func(watcher *fsnotify.Watcher) {
		_ = watcher.Close()
	}(watcher)

	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}

	curTime := time.Now()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// Skip Write Event happen in 200ms window
				if time.Since(curTime) < 200*time.Millisecond {
					continue
				}
				curTime = time.Now()

				html, err := RenderMarkdown(path)
				if err != nil {
					log.Println("Failed to re-render markdown:", err)
					continue
				}
				lock.Lock()
				rotIdx = (rotIdx + 1) % 2
				outPath, err := writeRotatingHTML(path, []byte(html), rotIdx)
				lock.Unlock()

				if err != nil {
					log.Println("Failled to write HTML:", err)
				}

				// Dispatch is needed because this is a go routine, not a main thread
				cWB.Dispatch(func() {
					// Save scroll position + Open
					cWB.Eval(fmt.Sprintf("(async function(){await SaveScroll(window.scrollY);window.location.href=%q})()", "file://"+outPath))
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}
