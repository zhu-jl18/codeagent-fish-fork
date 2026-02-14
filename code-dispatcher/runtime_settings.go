package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type runtimeSettingsOverrideState struct {
	mu      sync.RWMutex
	enabled bool
	values  map[string]string
}

var runtimeSettingsOverride runtimeSettingsOverrideState

func resolveDispatcherHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ""
	}
	return filepath.Join(home, ".code-dispatcher")
}

func resolveDispatcherEnvFile() string {
	base := resolveDispatcherHomeDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, ".env")
}

func resolvePromptBaseDir() string {
	base := resolveDispatcherHomeDir()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "prompts")
}

func loadRuntimeSettingsFile(path string) map[string]string {
	path = strings.TrimSpace(path)
	if path == "" {
		return map[string]string{}
	}

	file, err := os.Open(path)
	if err != nil {
		return map[string]string{}
	}
	defer file.Close()

	settings := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}
		idx := strings.IndexByte(line, '=')
		if idx <= 0 {
			continue
		}

		key := strings.TrimSpace(line[:idx])
		if key == "" {
			continue
		}
		value := strings.TrimSpace(line[idx+1:])
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		settings[key] = value
	}

	return settings
}

func lookupRuntimeSetting(key string) (string, bool) {
	runtimeSettingsOverride.mu.RLock()
	if runtimeSettingsOverride.enabled {
		value, ok := runtimeSettingsOverride.values[key]
		runtimeSettingsOverride.mu.RUnlock()
		if !ok {
			return "", false
		}
		return strings.TrimSpace(value), true
	}
	runtimeSettingsOverride.mu.RUnlock()

	settings := loadRuntimeSettingsFile(resolveDispatcherEnvFile())
	value, ok := settings[key]
	if !ok {
		return "", false
	}
	return strings.TrimSpace(value), true
}

func getEnv(key, defaultValue string) string {
	if val, ok := lookupRuntimeSetting(key); ok && val != "" {
		return val
	}
	return defaultValue
}

func runtimeInjectedEnv() map[string]string {
	runtimeSettingsOverride.mu.RLock()
	if runtimeSettingsOverride.enabled {
		out := make(map[string]string)
		for key, value := range runtimeSettingsOverride.values {
			if isDispatcherControlKey(key) {
				continue
			}
			if strings.TrimSpace(key) == "" {
				continue
			}
			out[key] = value
		}
		runtimeSettingsOverride.mu.RUnlock()
		if len(out) == 0 {
			return nil
		}
		return out
	}
	runtimeSettingsOverride.mu.RUnlock()

	settings := loadRuntimeSettingsFile(resolveDispatcherEnvFile())
	if len(settings) == 0 {
		return nil
	}

	out := make(map[string]string)
	for key, value := range settings {
		if isDispatcherControlKey(key) {
			continue
		}
		if strings.TrimSpace(key) == "" {
			continue
		}
		out[key] = value
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func isDispatcherControlKey(key string) bool {
	return strings.HasPrefix(key, "CODE_DISPATCHER_")
}

func setRuntimeSettingsForTest(values map[string]string) {
	runtimeSettingsOverride.mu.Lock()
	defer runtimeSettingsOverride.mu.Unlock()

	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	runtimeSettingsOverride.values = clone
	runtimeSettingsOverride.enabled = true
}

func resetRuntimeSettingsForTest() {
	runtimeSettingsOverride.mu.Lock()
	defer runtimeSettingsOverride.mu.Unlock()

	runtimeSettingsOverride.enabled = false
	runtimeSettingsOverride.values = nil
}
