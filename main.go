package main

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/auribuo/userdir"
	"github.com/charmbracelet/log"
	"github.com/fsnotify/fsnotify"
)

func runHooks(hookDir string) {
	if _, err := os.Stat(hookDir); err != nil {
		log.Warn("Hook dir does not exist", "dir", hookDir)
		return
	}

	entries, err := os.ReadDir(hookDir)
	if err != nil {
		log.Error("Failed to read hook dir", "dir", hookDir, "reason", err) 
		return
	}
	for _, hook := range entries {
		if hook.IsDir() {
			log.Warn("Found directory, ignoring", "file", hook.Name())
		}
		proc := exec.Command("sh", filepath.Join(hookDir, hook.Name()))
		proc.Env = os.Environ()
		err := proc.Run()
		if err != nil {
			log.Error("Hook failed to run", "hook", hook.Name(), "reason", err)
		}
	}
}

func main() {
	log.SetLevel(log.DebugLevel)
	cfgDir, err := userdir.Base().Config()
	if err != nil {
		log.Fatal("Failed to get config directory", "reason", err)
	}
	log.Debug("Found config directory", "dir", cfgDir)

	homeDir, err := userdir.Base().Home()
	if err != nil {
		log.Fatal("Failed to get home directory", "reason", err)
	}
	log.Debug("Found home directory", "dir", cfgDir)
	caelestiaPath := filepath.Join(homeDir, ".local", "state", "caelestia")
	caelestiaHookPath := filepath.Join(cfgDir, "caelestia", "hooks")

	watcher, err := fsnotify.NewWatcher()
	watcher.Add(caelestiaPath)
	if err != nil {
		log.Fatal("Failed to create watcher", "reason", err)
	}
	defer watcher.Close()
	log.Info("Starting watching for theme changes", "path", caelestiaPath)

	for event := range watcher.Events {
		log.Debug("Got event", "name", event.Name, "op", event.Op)
		if (event.Has(fsnotify.Remove)) && filepath.Base(event.Name) == "theme.lock" {
			log.Info("Detected theme change, running hooks")
			runHooks(caelestiaHookPath)
		}
	}
}
