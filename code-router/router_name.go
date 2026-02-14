package main

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultRouterName = "code-router"
)

var executablePathFn = os.Executable

func normalizeRouterName(path string) string {
	if path == "" {
		return ""
	}

	base := filepath.Base(path)
	base = strings.TrimSuffix(base, ".exe") // tolerate Windows executables

	if base == defaultRouterName {
		return base
	}

	return ""
}

// currentRouterName resolves the router name based on the invoked binary.
// Only known names are honored to avoid leaking build/test binary names into logs.
func currentRouterName() string {
	if len(os.Args) == 0 {
		return defaultRouterName
	}

	if name := normalizeRouterName(os.Args[0]); name != "" {
		return name
	}

	execPath, err := executablePathFn()
	if err == nil {
		if name := normalizeRouterName(execPath); name != "" {
			return name
		}

		if resolved, err := filepath.EvalSymlinks(execPath); err == nil {
			if name := normalizeRouterName(resolved); name != "" {
				return name
			}
		}
	}

	return defaultRouterName
}

// logPrefixes returns the set of accepted log name prefixes, including the
// current router name and legacy aliases.
func logPrefixes() []string {
	prefixes := []string{currentRouterName(), defaultRouterName}
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
// Defaults to the current router name when available, otherwise falls back
// to the canonical default name.
func primaryLogPrefix() string {
	prefixes := logPrefixes()
	if len(prefixes) == 0 {
		return defaultRouterName
	}
	return prefixes[0]
}
