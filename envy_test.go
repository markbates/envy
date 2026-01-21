package envy

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_FromSlice(t *testing.T) {
	t.Parallel()

	good := []string{"KEY1=VALUE1", "KEY2=VALUE2"}
	bad := []string{
		"MALFORMEDENTRY",
		"KEY2=VALUE2",
		"KEY1=VALUE1",
		"=NOVALUE",
		"   ",
		"KEY3=valueWith=equals",
		"// comment",
		"",
		"KEY4=KEY5-=baz",
	}

	tcs := []struct {
		name  string
		input []string
		exp   []string
	}{
		{
			name:  "well formed entries",
			input: good,
			exp:   good,
		},
		{
			name:  "malformed entries ignored",
			input: bad,
			exp:   []string{"KEY1=VALUE1", "KEY2=VALUE2", "KEY3=valueWith=equals", "KEY4=KEY5-=baz"},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := require.New(t)
			got := FromSlice(tc.input)
			r.Equal(tc.exp, got.Environ())
		})
	}

}

func Test_FromMap(t *testing.T) {
	t.Parallel()

	exp := []string{"KEY1=VALUE1", "KEY2=VALUE2"}

	tcs := []struct {
		name  string
		input map[string]string
		exp   []string
	}{
		{
			name:  "well formed map",
			input: map[string]string{"KEY1": "VALUE1", "KEY2": "VALUE2"},
			exp:   exp,
		},
		{
			name: "malformed map",
			input: map[string]string{
				"":          "NOVALUE",
				"KEY2":      "VALUE2",
				"KEY1":      "VALUE1",
				"KEY3":      "valueWith=equals",
				"//foo":     "bar",
				"   ":       "blank",
				"KEY4=KEY5": "baz",
			},
			exp: append(exp, "KEY3=valueWith=equals"),
		},
		{
			name:  "empty map",
			input: map[string]string{},
			exp:   []string{},
		},
		{
			name:  "nil map",
			input: nil,
			exp:   []string{},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			got := FromMap(tc.input)
			r.Equal(tc.exp, got.Environ())
		})
	}
}

func Test_Zero(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	e := Zero()
	r.NotNil(e)
	r.Empty(e.Environ())
}

func Test_New(t *testing.T) {
	t.Parallel()
	r := require.New(t)

	e := New()
	r.NotNil(e)
	r.NotEmpty(e.Environ())
	r.Equal(e.Getenv("PATH"), os.Getenv("PATH"))

}

type brokenReader struct{}

func (b *brokenReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

func Test_FromReader(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name  string
		input io.Reader
		exp   []string
		err   bool
	}{
		{
			name:  "empty reader",
			input: strings.NewReader(""),
			exp:   []string{},
		},
		{
			name:  "single entry",
			input: strings.NewReader("KEY1=VALUE1;"),
			exp:   []string{"KEY1=VALUE1"},
		},
		{
			name:  "multiple entries",
			input: strings.NewReader("KEY1=VALUE1;KEY2=VALUE2;KEY3=VALUE3;"),
			exp:   []string{"KEY1=VALUE1", "KEY2=VALUE2", "KEY3=VALUE3"},
		},
		{
			name:  "entries with spaces and newlines",
			input: strings.NewReader("  KEY1=VALUE1; \nKEY2=VALUE2;\n\n KEY3=VALUE3;  "),
			exp:   []string{"KEY1=VALUE1", "KEY2=VALUE2", "KEY3=VALUE3"},
		},
		{
			name:  "buffer error",
			input: bytes.NewReader([]byte{0xff, 0xfe, 0xfd}), // invalid UTF-8
			exp:   []string{},
		},
		{
			name:  "nil reader",
			input: nil,
			err:   true,
		},
		{
			name:  "broken reader",
			input: &brokenReader{},
			err:   true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)

			env, err := FromReader(tc.input, ';')
			if tc.err {
				r.Error(err)
				return
			}

			r.NoError(err)

			got := env.Environ()
			r.Equal(tc.exp, got)
		})
	}
}

func Test_FromFile(t *testing.T) {
	t.Parallel()

	td := os.DirFS("testdata")

	tcs := []struct {
		name string
		path string
		cab  fs.FS
		exp  []string
		err  bool
	}{
		{
			name: "valid env file",
			path: "valid.env",
			cab:  td,
			exp:  []string{"KEY1=VALUE1", "KEY2=VALUE2"},
		},
		{
			name: "non-existent file",
			path: "nonexistent.env",
			cab:  td,
			err:  true,
		},
		{
			name: "mixed env file",
			path: "mixed.env",
			cab:  td,
			exp:  []string{"KEY1=KEY1", "KEY2=VALUE2", "KEY3=VALUE3"},
		},
		{
			name: "nil fs",
			path: "nil.env",
			cab:  nil,
			err:  true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)

			env, err := FromFile(tc.cab, tc.path)
			if tc.err {
				r.Error(err)
				return
			}

			r.NoError(err)

			got := env.Environ()
			r.Equal(tc.exp, got)
		})
	}
}

func Test_With(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name string
		env  *Env
		fn   func() (*Env, error)
		exp  []string
		err  bool
	}{
		{
			name: "add new variable",
			env:  FromMap(map[string]string{"KEY1": "VALUE1"}),
			fn: func() (*Env, error) {
				return FromMap(map[string]string{"KEY2": "VALUE2"}), nil
			},
			exp: []string{"KEY1=VALUE1", "KEY2=VALUE2"},
		},
		{
			name: "modify existing variable",
			env:  FromMap(map[string]string{"KEY1": "VALUE1"}),
			fn: func() (*Env, error) {
				return FromMap(map[string]string{"KEY1": "NEWVALUE1"}), nil
			},
			exp: []string{"KEY1=NEWVALUE1"},
		},
		{
			name: "nil env",
			env:  nil,
			fn: func() (*Env, error) {
				return FromMap(map[string]string{"KEY": "VALUE"}), nil
			},
			err: true,
		},
		{
			name: "fn returns error",
			env:  FromMap(map[string]string{"KEY1": "VALUE1"}),
			fn: func() (*Env, error) {
				return nil, fmt.Errorf("function error")
			},
			err: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			merged, err := With(tc.env, tc.fn)
			if tc.err {
				r.Error(err)
				return
			}

			r.NoError(err)
			got := merged.Environ()
			r.Equal(tc.exp, got)
		})
	}

}

// ExampleWith demonstrates how to use the With function to create a modified
// environment based on an existing one. It shows how to merge changes and load
// additional variables from a file.
func ExampleWith() {
	base := New() // starts with the current environment

	// modified history is:
	// 1. os.Environ() -> base
	// 2. base + custom PATH -> first With
	// 3. first With + valid.env file -> second With
	modified, err := With(base, func() (*Env, error) {
		// Create a new Env with the modifications
		e := FromMap(map[string]string{
			"PATH": "/custom/path",
		})

		// Merge the modifications into the base Env
		e, err := base.Merge(e)
		if err != nil {
			return nil, err
		}

		// modify further by loading from a file
		return With(e, func() (*Env, error) {
			return FromFile(os.DirFS("testdata"), "valid.env")
		})
	})

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("PATH:", modified.Getenv("PATH"))
	fmt.Println("KEY1:", modified.Getenv("KEY1"))

	// Output:
	// PATH: /custom/path
	// KEY1: VALUE1

}
