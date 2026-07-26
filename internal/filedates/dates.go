// Package filedates changes creation and modification timestamps of files.
package filedates

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Options describes which timestamps have to be applied.
type Options struct {
	Created     time.Time
	SetCreated  bool
	Modified    time.Time
	SetModified bool
}

// Result reports the outcome for a single processed file.
type Result struct {
	Path string
	Err  error
}

// Apply walks every target (file or directory) and applies the timestamps.
// Directories are processed recursively, including their subdirectories.
func Apply(targets []string, opts Options) []Result {
	var results []Result
	for _, target := range Expand(targets) {
		results = append(results, Result{Path: target, Err: applyOne(target, opts)})
	}
	return results
}

// Expand resolves the given paths into a flat list of files, walking
// directories recursively. Directories themselves are included so their own
// timestamps are updated too.
func Expand(targets []string) []string {
	var out []string
	seen := make(map[string]bool)
	add := func(p string) {
		if abs, err := filepath.Abs(p); err == nil {
			p = abs
		}
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	for _, target := range targets {
		info, err := os.Stat(target)
		if err != nil {
			add(target)
			continue
		}
		if !info.IsDir() {
			add(target)
			continue
		}
		_ = filepath.WalkDir(target, func(path string, _ fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			add(path)
			return nil
		})
	}
	return out
}

func applyOne(path string, opts Options) error {
	if opts.SetModified {
		if err := os.Chtimes(path, opts.Modified, opts.Modified); err != nil {
			return err
		}
	}
	if opts.SetCreated {
		if err := setCreationTime(path, opts.Created); err != nil {
			return err
		}
	}
	return nil
}
