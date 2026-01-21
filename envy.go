// Package envy provides a concurrency-safe, in-memory view of environment
// variables that mirrors the behavior of the standard library's environment
// helpers without mutating the process state.
package envy

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
)

// Zero returns a new Env with no environment variables set. It is useful when
// you want a clean slate that is completely detached from the process
// environment.
func Zero() *Env {
	return FromMap(map[string]string{})
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
	em := map[string]string{}
	for _, env := range envs {
		// trim spaces
		env = strings.TrimSpace(env)

		// split into key and value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		// set key/value
		em[parts[0]] = parts[1]
	}

	return FromMap(em)
}

// FromMap wraps the provided map in a new Env. If the map is nil, an empty map
// is created. The map is used as-is (not copied), so callers should provide a
// map they own when sharing an Env between components.
func FromMap(envs map[string]string) *Env {
	if envs == nil {
		envs = map[string]string{}
	}

	for k := range envs {
		// trim spaces
		s := strings.TrimSpace(k)

		// ignore empty keys, comments, and keys with '='
		if s == "" || strings.Contains(s, "=") || strings.HasPrefix(s, "//") {
			delete(envs, k)
			continue
		}
	}

	return &Env{
		envs: envs,
	}
}

// FromReader reads environment entries from r, splitting on sep and trimming
// surrounding whitespace. It returns an error for a nil reader or scanner
// failures (including invalid UTF-8).
func FromReader(r io.Reader, sep byte) (*Env, error) {
	if r == nil {
		return nil, fmt.Errorf("nil reader")
	}

	buf := bufio.NewScanner(r)

	buf.Split(func(data []byte, eof bool) (int, []byte, error) {
		// trim space
		tsd := func(b []byte) []byte {
			return bytes.TrimSpace(b)
		}

		// return if no data
		if len(data) == 0 {
			return 0, nil, nil
		}

		// split on sep
		if i := bytes.IndexByte(data, sep); i >= 0 {
			return i + 1, tsd(data[0:i]), nil
		}

		// handle eof
		if eof {
			return len(data), tsd(data), nil
		}

		return 0, nil, nil
	})

	envs := []string{}
	for buf.Scan() {
		envs = append(envs, buf.Text())
	}

	if err := buf.Err(); err != nil {
		return nil, err
	}

	return FromSlice(envs), nil
}

// FromFile reads newline-separated environment entries from the provided
// filesystem path. It returns an error for a nil fs.FS or any read failure.
func FromFile(cab fs.FS, path string) (e *Env, err error) {
	if cab == nil {
		return nil, fmt.Errorf("nil fs.FS")
	}

	f, err := cab.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		cerr := f.Close()
		if cerr == nil {
			return
		}

		if err == nil {
			err = cerr
			return
		}
		err = errors.Join(err, cerr)
	}()

	lines := []string{}
	buf := bufio.NewScanner(f)
	for buf.Scan() {
		lines = append(lines, buf.Text())
	}

	if err := buf.Err(); err != nil {
		return nil, err
	}

	return FromSlice(lines), nil
}

// With calls fn to produce an Env and merges the result into env. It returns
// an error if env is nil, if fn fails, or if the merge fails.
func With(env *Env, fn func() (*Env, error)) (*Env, error) {
	if env == nil {
		return nil, fmt.Errorf("nil env")
	}

	n, err := fn()
	if err != nil {
		return nil, err
	}

	return env.Merge(n)
}
