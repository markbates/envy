// Package envy provides a concurrency-safe, in-memory view of environment
// variables that mirrors the behavior of the standard library's environment
// helpers without mutating the process state.
package envy

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

// Zero returns a new Env with no environment variables set. It is useful when
// you want a clean slate that is completely detached from the process
// environment.
func Zero() *Env {
	return &Env{
		envs: map[string]string{},
	}
}

// New returns an Env populated with the current process's environment
// variables. Future calls to Setenv/Unsetenv modify the Env only and do not
// change the process environment.
func New() *Env {
	return FromSlice(os.Environ())
}

// FromSlice builds an Env from a slice of strings in the form "KEY=VALUE".
// Malformed entries are ignored. Later entries with the same key overwrite
// earlier ones, matching the standard environment semantics.
func FromSlice(envs []string) *Env {
	envMap := map[string]string{}
	for _, env := range envs {
		var key, value string
		n, _ := fmt.Sscanf(env, "%[^=]=%s", &key, &value)
		if n == 2 {
			envMap[key] = value
		}
	}
	return FromMap(envMap)
}

// FromMap wraps the provided map in a new Env. If the map is nil, an empty map
// is created. The map is used as-is (not copied), so callers should provide a
// map they own when sharing an Env between components.
func FromMap(envs map[string]string) *Env {
	if envs == nil {
		envs = map[string]string{}
	}

	return &Env{
		envs: envs,
	}
}

// Env stores environment variables in memory with thread-safe access. A nil
// *Env is treated as empty and safe to read from, but mutating operations
// return an error.
type Env struct {
	// envs is a map that holds environment variables.
	envs map[string]string
	mu   sync.RWMutex
}

// Getenv returns the value of the environment variable named by key. It returns
// an empty string when the key is not present or the Env is nil, mirroring
// os.Getenv semantics.
func (e *Env) Getenv(key string) string {
	if e.IsNil() {
		return ""
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.envs[key]
}

// Setenv sets the value of the environment variable named by key. It returns an
// error if the Env or its backing map is nil.
func (e *Env) Setenv(key, value string) error {
	if e.IsNil() {
		return fmt.Errorf("nil env")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.envs[key] = value
	return nil
}

// Unsetenv deletes the environment variable named by key. Removing a missing
// key is a no-op. An error is returned if the Env or its backing map is nil.
func (e *Env) Unsetenv(key string) error {
	if e.IsNil() {
		return fmt.Errorf("nil env")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.envs, key)
	return nil
}

// IsNil reports whether the receiver is nil or its underlying map is nil. This
// allows callers to safely check Env values that may not have been
// initialized.
func (e *Env) IsNil() bool {
	if e == nil {
		return true
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.envs == nil
}

// Environ returns a sorted slice of strings in the form "key=value" for every
// variable stored in the Env. The slice is deterministic to make comparisons in
// tests predictable.
func (e *Env) Environ() []string {
	if e.IsNil() {
		return []string{}
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	var envs []string
	for k, v := range e.envs {
		envs = append(envs, k+"="+v)
	}

	sort.Strings(envs)
	return envs
}

// Expandenv replaces ${var} or $var in the input string according to the
// stored environment variables. Unknown keys are replaced with the empty
// string. If the Env is nil, the input string is returned unchanged.
func (e *Env) Expandenv(s string) string {
	if e.IsNil() {
		return s
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	return os.Expand(s, func(key string) string {
		if val, ok := e.envs[key]; ok {
			return val
		}
		return ""
	})
}
