package main

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

type EventType int

const (
	ProcessStarted EventType = iota
	ProcessStopped
)

type ProcessEvent struct {
	Type EventType `json:"type"`
	PID  int32     `json:"pid"`
	Name string    `json:"name"`
}

type ProcessWatcher struct {
	Interval time.Duration
}

var xxmiScanDirs = [4]string{
	os.Getenv("APPDATA"),
	os.Getenv("LOCALAPPDATA"),
	os.Getenv("ProgramFiles"),
	os.Getenv("ProgramFiles(x86)"),
}

func getXXMILocation() (string, bool) {
	for _, dir := range xxmiScanDirs {
		if dir == "" {
			continue
		}

		exePath, err := searchExecutable(dir, "XXMI Launcher.exe")
		if err != nil {
			continue
		}

		dir := filepath.Dir(exePath)
		for range 2 {
			dir = filepath.Dir(dir)
		}

		return dir, true
	}

	return "", false
}

func searchExecutable(root, name string) (string, error) {
	var found string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if !d.IsDir() && strings.EqualFold(d.Name(), name) {
			found = path
			return fs.SkipAll
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if found == "" {
		return "", os.ErrNotExist
	}

	return found, nil
}

func NewProcessWatcher(interval time.Duration) *ProcessWatcher {
	if interval <= 0 {
		interval = time.Second
	}

	return &ProcessWatcher{
		Interval: interval,
	}
}

func (w *ProcessWatcher) Current(names ...string) map[string]ProcessEvent {
	watched := make(map[string]bool)

	for _, name := range names {
		watched[strings.ToLower(name)] = true
	}

	events := make(chan ProcessEvent)

	go func() {
		w.Check(watched, make(map[int32]string), events)
		close(events)
	}()

	curr := make(map[string]ProcessEvent)
	for {
		event, ok := <-events
		if !ok {
			break
		}
		curr[event.Name] = event
	}

	return curr
}

func (w *ProcessWatcher) Watch(ctx context.Context, names ...string) <-chan ProcessEvent {

	events := make(chan ProcessEvent)
	watched := make(map[string]bool)

	for _, name := range names {
		watched[strings.ToLower(name)] = true
	}

	go func() {
		defer close(events)

		ticker := time.NewTicker(w.Interval)
		defer ticker.Stop()

		previous := w.Check(watched, make(map[int32]string), events)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				previous = w.Check(watched, previous, events)
			}
		}
	}()

	return events
}

func (w *ProcessWatcher) Check(watched map[string]bool, previous map[int32]string, events chan ProcessEvent) map[int32]string {
	processes, err := process.Processes()
	if err != nil {
		return previous
	}

	current := make(map[int32]string)

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		name = strings.ToLower(name)

		if !watched[name] {
			continue
		}

		if name == "client-win64-shipping.exe" &&
			!isGenericUnityGameWutheringWaves(p) {
			continue
		}

		current[p.Pid] = name

		if _, exists := previous[p.Pid]; !exists {
			events <- ProcessEvent{
				Type: ProcessStarted,
				PID:  p.Pid,
				Name: name,
			}
		}
	}

	for pid, name := range previous {
		if _, ok := current[pid]; !ok {
			events <- ProcessEvent{
				Type: ProcessStopped,
				PID:  pid,
				Name: name,
			}
		}
	}

	return current
}

func isGenericUnityGameWutheringWaves(p *process.Process) bool {
	exe, err := p.Exe()
	if err != nil {
		return false
	}

	exe, err = filepath.Abs(exe)
	if err != nil {
		return false
	}

	return strings.Contains(
		strings.ToLower(exe),
		"wuthering",
	)
}
