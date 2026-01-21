package envy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Env_Gentenv(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name string
		env  *Env
		exp  string
	}{
		{
			name: "nil env",
			env:  nil,
			exp:  "",
		},
		{
			name: "empty env",
			env:  Zero(),
			exp:  "",
		},
		{
			name: "found key",
			env:  FromMap(map[string]string{"KEY": "VALUE"}),
			exp:  "VALUE",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			got := tc.env.Getenv("KEY")
			r.Equal(tc.exp, got)
		})
	}

}

func Test_Env_Setenv(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name  string
		env   *Env
		key   string
		value string
		err   bool
	}{
		{
			name:  "nil env",
			env:   nil,
			key:   "KEY",
			value: "VALUE",
			err:   true,
		},
		{
			name:  "set key",
			env:   Zero(),
			key:   "KEY",
			value: "VALUE",
		},
		{
			name: "nil map",
			env: &Env{
				envs: nil,
			},
			key:   "KEY",
			value: "VALUE",
			err:   true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			err := tc.env.Setenv(tc.key, tc.value)
			if tc.err {
				r.Error(err)
				return
			}

			r.NoError(err)
			got := tc.env.Getenv(tc.key)
			r.Equal(tc.value, got)
		})
	}
}

func Test_Env_Unsetenv(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name string
		env  *Env
		key  string
		err  bool
	}{
		{
			name: "nil env",
			env:  nil,
			key:  "KEY",
			err:  true,
		},
		{
			name: "unset existing key",
			env:  FromMap(map[string]string{"KEY": "VALUE"}),
			key:  "KEY",
		},
		{
			name: "unset missing key",
			env:  FromMap(map[string]string{}),
			key:  "KEY",
		},
		{
			name: "nil map",
			env: &Env{
				envs: nil,
			},
			key: "KEY",
			err: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			err := tc.env.Unsetenv(tc.key)
			if tc.err {
				r.Error(err)
				return
			}

			r.NoError(err)
			got := tc.env.Getenv(tc.key)
			r.Equal("", got)
		})
	}
}

func Test_Env_IsNil(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name string
		env  *Env
		exp  bool
	}{
		{
			name: "nil env",
			env:  nil,
			exp:  true,
		},
		{
			name: "nil map",
			env: &Env{
				envs: nil,
			},
			exp: true,
		},
		{
			name: "non-nil env",
			env:  FromMap(map[string]string{}),
			exp:  false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			got := tc.env.IsNil()
			r.Equal(tc.exp, got)
		})
	}
}

func Test_Env_Environ(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name  string
		input *Env
		exp   []string
	}{
		{
			name:  "empty env",
			input: Zero(),
			exp:   []string{},
		},
		{
			name:  "populated env",
			input: FromMap(map[string]string{"KEY2": "VALUE2", "KEY1": "VALUE1"}),
			exp:   []string{"KEY1=VALUE1", "KEY2=VALUE2"},
		},
		{
			name: "nil map",
			input: &Env{
				envs: nil,
			},
			exp: []string{},
		},
		{
			name:  "nil env",
			input: nil,
			exp:   []string{},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			got := tc.input.Environ()
			r.Equal(tc.exp, got)
		})
	}
}

func Test_Env_Expandenv(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name  string
		env   *Env
		input string
		exp   string
	}{
		{
			name:  "nil env",
			env:   nil,
			input: "Value is $KEY",
			exp:   "Value is $KEY",
		},
		{
			name:  "empty env",
			env:   Zero(),
			input: "Value is $KEY",
			exp:   "Value is ",
		},
		{
			name:  "key found",
			env:   FromMap(map[string]string{"KEY": "VALUE"}),
			input: "Value is $KEY",
			exp:   "Value is VALUE",
		},
		{
			name:  "key not found",
			env:   FromMap(map[string]string{}),
			input: "Value is $KEY",
			exp:   "Value is ",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			got := tc.env.Expandenv(tc.input)
			r.Equal(tc.exp, got)
		})
	}
}

func Test_Env_Merge(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name string
		env1 *Env
		env2 *Env
		exp  []string
		err  bool
	}{
		{
			name: "merge two populated envs",
			env1: FromMap(map[string]string{"KEY1": "VALUE1", "KEY2": "VALUE2"}),
			env2: FromMap(map[string]string{"KEY2": "NEWVALUE2", "KEY3": "VALUE3"}),
			exp:  []string{"KEY1=VALUE1", "KEY2=NEWVALUE2", "KEY3=VALUE3"},
		},
		{
			name: "merge into nil env",
			env1: nil,
			env2: FromMap(map[string]string{"KEY": "VALUE"}),
			err:  true,
		},
		{
			name: "merge from nil env",
			env1: FromMap(map[string]string{"KEY": "VALUE"}),
			env2: nil,
			err:  true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			merged, err := tc.env1.Merge(tc.env2)
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

func Test_Env_IsSet(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name string
		env  *Env
		key  string
		exp  bool
	}{
		{
			name: "nil env",
			env:  nil,
			key:  "KEY",
			exp:  false,
		},
		{
			name: "key is set",
			env:  FromMap(map[string]string{"KEY": "VALUE"}),
			key:  "KEY",
			exp:  true,
		},
		{
			name: "key is not set",
			env:  FromMap(map[string]string{}),
			key:  "KEY",
			exp:  false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)
			got := tc.env.IsSet(tc.key)
			r.Equal(tc.exp, got)
		})
	}
}
