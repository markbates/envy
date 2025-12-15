package envy

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Env(t *testing.T) {
	t.Parallel()

	r := require.New(t)
	e := Zero()
	r.False(e.IsNil())
	r.Equal("", e.Getenv("PATH"))

	const pexp = "/usr/local/bin"

	tcs := []struct {
		name string
		e    *Env
		pexp string
	}{
		{
			name: "Zero Env",
			e:    Zero(),
			pexp: "",
		},
		{
			name: "FromMap Env",
			e:    FromMap(map[string]string{"PATH": pexp}),
			pexp: pexp,
		},
		{
			name: "FromMap Env with nil",
			e:    FromMap(nil),
			pexp: "",
		},
		{
			name: "FromSlice Env",
			e:    FromSlice([]string{fmt.Sprintf("PATH=%s", pexp)}),
			pexp: pexp,
		},
		{
			name: "New Env",
			e:    New(),
			pexp: os.Getenv("PATH"),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := require.New(t)

			err := tc.e.Setenv("PATH", tc.pexp)
			r.NoError(err)
			r.Equal(tc.pexp, tc.e.Getenv("PATH"))

			err = tc.e.Setenv("FOO", "BAR")
			r.NoError(err)
			r.Equal([]string{"FOO=BAR", fmt.Sprintf("PATH=%s", tc.pexp)}, tc.e.Environ())
			r.Equal("I am BAR", tc.e.Expandenv("I am ${FOO}"))

			err = tc.e.Unsetenv("FOO")
			r.NoError(err)
			r.Equal("", tc.e.Getenv("FOO"))
			r.Equal("I am ", tc.e.Expandenv("I am ${FOO}"))

		})
	}

	t.Run("Nil Env", func(t *testing.T) {
		t.Parallel()
		var e *Env
		r := require.New(t)
		r.True(e.IsNil())
		r.Equal("", e.Getenv("PATH"))

		err := e.Setenv("PATH", pexp)
		r.Error(err)
		r.Equal("", e.Getenv("PATH"))

		err = e.Unsetenv("PATH")
		r.Error(err)

		r.Equal([]string{}, e.Environ())
		r.Equal("I am ${FOO}", e.Expandenv("I am ${FOO}"))
	})
}

// func Test_ZeroEnv(t *testing.T) {
// 	t.Parallel()
// 	r := require.New(t)

// 	e := Zero()
// 	r.NotNil(e)

// 	exps := map[string]string{
// 		"PATH": "",
// 	}
// 	env_tests(t, e, exps)
// }

// func Test_NilEnv(t *testing.T) {
// 	t.Parallel()
// 	r := require.New(t)

// 	var e *Env
// 	r.Nil(e)

// 	exps := map[string]string{
// 		"PATH": "",
// 	}
// 	env_tests(t, e, exps)

// }

// func Test_FromMap(t *testing.T) {
// 	t.Parallel()
// 	r := require.New(t)

// 	exps := map[string]string{
// 		"KEY1": "VALUE1",
// 		"KEY2": "VALUE2",
// 		"PATH": "",
// 	}

// 	e := FromMap(exps)
// 	r.NotNil(e)
// 	env_tests(t, e, exps)

// 	t.Run("FromMap with nil", func(t *testing.T) {
// 		t.Parallel()
// 		e := FromMap(nil)
// 		r.NotNil(e)

// 		exps := map[string]string{}
// 		env_tests(t, e, exps)
// 	})
// }

// func env_tests(t testing.TB, e *Env, expects map[string]string) {
// 	r := require.New(t)

// 	for k, v := range expects {
// 		r.Equal(v, e.Getenv(k))
// 	}

// 	for k := range expects {
// 		err := e.Unsetenv(k)
// 		r.NoError(err)
// 		r.Equal("", e.Getenv(k))
// 	}

// 	err := e.Setenv("NEW_KEY", "NEW_VALUE")
// 	r.NoError(err)
// 	r.Equal("NEW_VALUE", e.Getenv("NEW_KEY"))
// }
