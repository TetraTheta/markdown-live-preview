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
	currentRotation int
	currentWebView  webview.WebView
	lock            sync.Mutex
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

	rotationPath, err := writeRotatingHTML(path, []byte(html), 0)
	if err != nil {
		log.Fatalf("Failed to write HTML: %v", err)
	}

	w := webview.New(true)
	currentWebView = w
	defer w.Destroy()

	w.SetTitle("Markdown Live Preview")
	w.SetSize(800, 600, webview.HintNone)
	w.Navigate("file://" + rotationPath)

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

	debounce := time.Now()

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				if time.Since(debounce) < 200*time.Millisecond {
					continue
				}
				debounce = time.Now()

				html, err := RenderMarkdown(path)
				if err != nil {
					log.Println("Failed to re-render markdown:", err)
					continue
				}
				lock.Lock()
				currentRotation = (currentRotation + 1) % 2
				outPath, err := writeRotatingHTML(path, []byte(html), currentRotation)
				if err == nil {
					currentWebView.Dispatch(func() {
						currentWebView.Eval("scrollY = window.scrollY")
						currentWebView.Navigate("file://" + outPath)
						currentWebView.Eval("window.scrollTo(0, scrollY")
					})
				}
				lock.Unlock()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}
