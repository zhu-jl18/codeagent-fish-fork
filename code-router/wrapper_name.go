package main

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultWrapperName = "code-router"
)

var executablePathFn = os.Executable

func normalizeWrapperName(path string) string {
	if path == "" {
		return ""
	}

	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".exe") // tolerate Windows executables

	if base == defaultWrapperName {
		return base
	}

	return ""
}

// currentWrapperName resolves the wrapper name based on the invoked binary.
// Only known names are honored to avoid leaking build/test binary names into logs.
func currentWrapperName() string {
	if len(os.Args) == 0 {
		return defaultWrapperName
	}

	if name := normalizeWrapperName(os.Args[0]); name != "" {
		return name
	}

	execPath, err := executablePathFn()
	if err == nil {
		if name := normalizeWrapperName(execPath); name != "" {
			return name
		}

		if resolved, err := filepath.EvalSymlinks(execPath); err == nil {
			if name := normalizeWrapperName(resolved); name != "" {
				return name
			}
		}
	}

	return defaultWrapperName
}

// logPrefixes returns the set of accepted log name prefixes, including the
// current wrapper name and legacy aliases.
func logPrefixes() []string {
	prefixes := []string{currentWrapperName(), defaultWrapperName}
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
// Defaults to the current wrapper name when available, otherwise falls back
// to the canonical default name.
func primaryLogPrefix() string {
	prefixes := logPrefixes()
	if len(prefixes) == 0 {
		return defaultWrapperName
	}
	return prefixes[0]
}
