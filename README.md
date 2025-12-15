# envy

Small, thread-safe helpers for working with environment variables without mutating the process environment. `Env` wraps environment data in an in-memory map, making it easy to build isolated envs for tests or compose custom environments for commands.

## Features

- Thread-safe reads and writes via an internal RWMutex.
- Nil-safe: reading from a nil `*Env` returns empty values; mutating a nil `*Env` returns an error.
- Deterministic `Environ` output to keep tests stable.
- `Expandenv` replacement that uses the stored values rather than the process environment.
- Start from scratch with `Zero`, clone the current process with `New`, or build from maps/slices.

## Install

```bash
go get github.com/markbates/envy
```

## Quick start

```go
package main

import (
	"fmt"

	"github.com/markbates/envy"
)

func main() {
	env := envy.New() // copy the current process environment

	_ = env.Setenv("APP_ENV", "dev")

	fmt.Println(env.Getenv("PATH"))               // read like os.Getenv
	fmt.Println(env.Environ())                    // sorted slice of "key=value"
	fmt.Println(env.Expandenv("mode=${APP_ENV}")) // "mode=dev"
}
```

## Building isolated environments

- Use `envy.Zero()` for a blank environment that cannot leak the real process state.
- Convert existing data with `envy.FromSlice(os.Environ())` or `envy.FromMap(map[string]string{"FOO": "BAR"})`.
- `FromMap` uses the provided map directly; pass in a map you own when sharing an `Env` across components.

## Behavior and concurrency

- `Env` is safe for concurrent readers and writers.
- A nil `*Env` behaves as an empty environment for reads; `Setenv`/`Unsetenv` on a nil receiver return an error instead of panicking.
- `Environ` output is sorted for deterministic comparisons (handy in tests).
- `Expandenv` replaces `$var`/`${var}` placeholders with values stored in the `Env` and leaves unknown keys empty.

## Testing

Run the test suite with:

```bash
go test -v -cover -race ./...
```
