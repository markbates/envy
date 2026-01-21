package envy

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

// Env stores environment variables in memory with thread-safe access. A nil
// *Env is treated as empty and safe to read from, but mutating operations
// return an error. The zero value is not ready for mutation; use Zero, New, or
// FromMap to initialize it.
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

// IsNil reports whether the receiver or its backing map is nil. This allows
// callers to safely check Env values that may not have been initialized.
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

	envs := []string{}
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

// Merge returns a new Env containing the receiver's variables
// overridden by the variables from other. It returns an error
// if either Env is nil.
func (e *Env) Merge(other *Env) (*Env, error) {
	if e.IsNil() {
		return nil, fmt.Errorf("cannot merge into nil env")
	}

	if other.IsNil() {
		return nil, fmt.Errorf("cannot merge from nil env")
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	em := map[string]string{}
	for k, v := range e.envs {
		em[k] = v
	}

	for k, v := range other.envs {
		em[k] = v
	}

	return FromMap(em), nil
}

// IsSet reports whether key is present in the Env. It returns false for a nil Env.
func (e *Env) IsSet(key string) bool {
	if e.IsNil() {
		return false
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	_, ok := e.envs[key]
	return ok
}

func (e *Env) String() string {
	return strings.Join(e.Environ(), ";")
}
