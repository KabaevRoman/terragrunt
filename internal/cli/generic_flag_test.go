package cli_test

import (
	"errors"
	libflag "flag"
	"fmt"
	"io"
	"testing"

	"github.com/gruntwork-io/terragrunt/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericFlagStringApply(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		flag          cli.GenericFlag[string]
		args          []string
		envs          map[string]string
		expectedValue string
		expectedErr   error
	}{
		{
			cli.GenericFlag[string]{Name: "foo", EnvVars: []string{"FOO"}},
			[]string{"--foo", "arg-value"},
			map[string]string{"FOO": "env-value"},
			"arg-value",
			nil,
		},
		{
			cli.GenericFlag[string]{Name: "foo", EnvVars: []string{"FOO"}},
			nil,
			map[string]string{"FOO": "env-value"},
			"env-value",
			nil,
		},
		{
			cli.GenericFlag[string]{Name: "foo", EnvVars: []string{"FOO"}},
			nil,
			nil,
			"",
			nil,
		},
		{
			cli.GenericFlag[string]{Name: "foo", EnvVars: []string{"FOO"}, Destination: mockDestValue("default-value")},
			[]string{"--foo", "arg-value"},
			map[string]string{"FOO": "env-value"},
			"arg-value",
			nil,
		},
		{
			cli.GenericFlag[string]{Name: "foo", Destination: mockDestValue("default-value")},
			nil,
			nil,
			"default-value",
			nil,
		},
		{
			cli.GenericFlag[string]{Name: "foo", EnvVars: []string{"FOO"}},
			[]string{"--foo", "arg-value1", "--foo", "arg-value2"},
			nil,
			"",
			errors.New(`invalid value "arg-value2" for flag -foo: setting the flag multiple times`),
		},
	}

	for i, testCase := range testCases {
		testCase := testCase

		t.Run(fmt.Sprintf("testCase-%d", i), func(t *testing.T) {
			t.Parallel()

			testGenericFlagApply(t, &testCase.flag, testCase.args, testCase.envs, testCase.expectedValue, testCase.expectedErr)
		})
	}
}

func TestGenericFlagIntApply(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		flag          cli.GenericFlag[int]
		args          []string
		envs          map[string]string
		expectedValue int
		expectedErr   error
	}{
		{
			cli.GenericFlag[int]{Name: "foo", EnvVars: []string{"FOO"}},
			[]string{"--foo", "10"},
			map[string]string{"FOO": "20"},
			10,
			nil,
		},
		{
			cli.GenericFlag[int]{Name: "foo", EnvVars: []string{"FOO"}},
			[]string{},
			map[string]string{"FOO": "20"},
			20,
			nil,
		},
		{
			cli.GenericFlag[int]{Name: "foo", EnvVars: []string{"FOO"}},
			[]string{},
			map[string]string{"FOO": "monkey"},
			0,
			errors.New(`invalid value "monkey" for env var FOO: must be 32-bit integer`),
		},
		{
			cli.GenericFlag[int]{Name: "foo", Destination: mockDestValue(55)},
			nil,
			nil,
			55,
			nil,
		},
	}

	for i, testCase := range testCases {
		testCase := testCase

		t.Run(fmt.Sprintf("testCase-%d", i), func(t *testing.T) {
			t.Parallel()

			testGenericFlagApply(t, &testCase.flag, testCase.args, testCase.envs, testCase.expectedValue, testCase.expectedErr)
		})
	}
}

func TestGenericFlagInt64Apply(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		flag          cli.GenericFlag[int64]
		args          []string
		envs          map[string]string
		expectedValue int64
		expectedErr   error
	}{
		{
			cli.GenericFlag[int64]{Name: "foo", EnvVars: []string{"FOO"}},
			[]string{"--foo", "10"},
			map[string]string{"FOO": "20"},
			10,
			nil,
		},
		{
			cli.GenericFlag[int64]{Name: "foo", EnvVars: []string{"FOO"}},
			[]string{},
			map[string]string{"FOO": "20"},
			20,
			nil,
		},
		{
			cli.GenericFlag[int64]{Name: "foo", EnvVars: []string{"FOO"}},
			[]string{},
			map[string]string{"FOO": "monkey"},
			0,
			errors.New(`invalid value "monkey" for env var FOO: must be 64-bit integer`),
		},
		{
			cli.GenericFlag[int64]{Name: "foo", Destination: mockDestValue(int64(55))},
			nil,
			nil,
			55,
			nil,
		},
	}

	for i, testCase := range testCases {
		testCase := testCase

		t.Run(fmt.Sprintf("testCase-%d", i), func(t *testing.T) {
			t.Parallel()

			testGenericFlagApply(t, &testCase.flag, testCase.args, testCase.envs, testCase.expectedValue, testCase.expectedErr)
		})
	}
}

func testGenericFlagApply[T cli.GenericType](t *testing.T, flag *cli.GenericFlag[T], args []string, envs map[string]string, expectedValue T, expectedErr error) {
	t.Helper()

	var (
		actualValue          T
		expectedDefaultValue string
	)

	if flag.Destination == nil {
		flag.Destination = new(T)
	}

	expectedDefaultValue = fmt.Sprintf("%v", *flag.Destination)

	flag.LookupEnvFunc = func(key string) []string {
		if envs == nil {
			return nil
		}

		if val, ok := envs[key]; ok {
			return []string{val}
		}
		return nil
	}

	flagSet := libflag.NewFlagSet("test-cmd", libflag.ContinueOnError)
	flagSet.SetOutput(io.Discard)

	err := flag.Apply(flagSet)
	if err == nil {
		err = flagSet.Parse(args)
	}

	if expectedErr != nil {
		require.Error(t, err)
		require.ErrorContains(t, expectedErr, err.Error())
		return
	}
	require.NoError(t, err)

	actualValue = (flag.Value().Get()).(T)

	assert.Equal(t, expectedValue, actualValue)
	assert.Equal(t, fmt.Sprintf("%v", expectedValue), flag.GetValue(), "GetValue()")

	assert.Equal(t, len(args) > 0 || len(envs) > 0, flag.Value().IsSet(), "IsSet()")
	assert.Equal(t, expectedDefaultValue, flag.GetInitialTextValue(), "GetDefaultText()")

	assert.False(t, flag.Value().IsBoolFlag(), "IsBoolFlag()")
	assert.True(t, flag.TakesValue(), "TakesValue()")
}
