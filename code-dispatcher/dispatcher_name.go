package main

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultDispatcherName = "code-dispatcher"
)

var executablePathFn = os.Executable

func normalizeDispatcherName(path string) string {
	if path == "" {
		return ""
	}

	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".exe") // tolerate Windows executables

	if base == defaultDispatcherName {
		return base
	}

	return ""
}

// currentDispatcherName resolves the dispatcher name based on the invoked binary.
// Only known names are honored to avoid leaking build/test binary names into logs.
func currentDispatcherName() string {
	if len(os.Args) == 0 {
		return defaultDispatcherName
	}

	if name := normalizeDispatcherName(os.Args[0]); name != "" {
		return name
	}

	execPath, err := executablePathFn()
	if err == nil {
		if name := normalizeDispatcherName(execPath); name != "" {
			return name
		}

		if resolved, err := filepath.EvalSymlinks(execPath); err == nil {
			if name := normalizeDispatcherName(resolved); name != "" {
				return name
			}
		}
	}

	return defaultDispatcherName
}

// logPrefixes returns the set of accepted log name prefixes, including the
// current dispatcher name and legacy aliases.
func logPrefixes() []string {
	prefixes := []string{currentDispatcherName(), defaultDispatcherName}
	seen := make(map[string]struct{}, len(prefixes))
	var unique []string
	for _, prefix := range prefixes {
		if prefix == "" {
			continue
		}
		if _, ok := seen[prefix]; ok {
			continue
		}
		seen[prefix] = struct{}{}
		unique = append(unique, prefix)
	}
	return unique
}

// primaryLogPrefix returns the preferred filename prefix for log files.
// Defaults to the current dispatcher name when available, otherwise falls back
// to the canonical default name.
func primaryLogPrefix() string {
	prefixes := logPrefixes()
	if len(prefixes) == 0 {
		return defaultDispatcherName
	}
	return prefixes[0]
}
