package envy

import (
	"fmt"
	"os"
	"sort"
	"sync"
)

// Zero creates a new Env with no environment variables set.
func Zero() *Env {
	return &Env{
		envs: map[string]string{},
	}
}

// New creates a new Env populated with the current process's environment variables.
func New() *Env {
	return FromSlice(os.Environ())
}

// FromSlice creates a new Env from a slice of environment variable strings in the form "KEY=VALUE".
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

// FromMap creates a new Env from a map of environment variables.
func FromMap(envs map[string]string) *Env {
	if envs == nil {
		envs = map[string]string{}
	}

	return &Env{
		envs: envs,
	}
}

type Env struct {
	// envs is a map that holds environment variables.
	envs map[string]string
	mu   sync.RWMutex
}

func (e *Env) Getenv(key string) string {
	if e.IsNil() {
		return ""
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.envs[key]
}

func (e *Env) Setenv(key, value string) error {
	if e.IsNil() {
		return fmt.Errorf("nil env")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.envs[key] = value
	return nil
}

func (e *Env) Unsetenv(key string) error {
	if e.IsNil() {
		return fmt.Errorf("nil env")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.envs, key)
	return nil
}

func (e *Env) IsNil() bool {
	if e == nil {
		return true
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.envs == nil
}

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
