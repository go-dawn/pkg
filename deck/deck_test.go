package deck

import (
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"

	"github.com/stretchr/testify/assert"
)

func TestOsExit(t *testing.T) {
	var count int32
	SetupOsExit(func(code int) {
		atomic.AddInt32(&count, int32(code))
	})
	defer TeardownOsExit()

	OsExit(100)

	assert.Equal(t, int32(100), atomic.LoadInt32(&count))
}

func TestExecCommand(t *testing.T) {
	at := assert.New(t)

	t.Run("success", func(t *testing.T) {
		SetupCmd()
		defer TeardownCmd()

		cmd := ExecCommand("test", "success")

		b, err := cmd.CombinedOutput()
		at.Nil(err)
		at.Equal("[test success]", string(b))
	})

	t.Run("error", func(t *testing.T) {
		SetupCmdError()
		defer TeardownCmd()

		cmd := ExecCommand("test", "error")

		b, err := cmd.CombinedOutput()
		at.NotNil(err)
		at.Equal("", string(b))
		at.Contains(err.Error(), errorCommand)
	})

	t.Run("stderr", func(t *testing.T) {
		SetupCmdStderr()
		defer TeardownCmd()

		cmd := ExecCommand("test", "stderr")

		b, err := cmd.CombinedOutput()
		at.NotNil(err)
		at.Equal("[test stderr]", string(b))
	})
}

func TestHelperCommand(t *testing.T) {
	HandleCommand(func(args []string, expectStderr bool) {
		if expectStderr {
			_, _ = fmt.Fprintf(os.Stderr, "%v", args)
			os.Exit(1)
		}
		_, _ = fmt.Fprintf(os.Stdout, "%v", args)
	})
}

func TestHandleCommand(t *testing.T) {
	at := assert.New(t)

	SetupEnvs(Envs{"GO_WANT_HELPER_COMMAND": "1"})
	defer TeardownEnvs()

	SetupOsExit()
	defer TeardownOsExit()

	os.Args = append(os.Args, "--", "test")

	HandleCommand(func(args []string, expectStderr bool) {
		at.Equal(1, len(args))
		at.Equal("test", args[0])
		at.False(expectStderr)
	})
}

func TestExecLookPath(t *testing.T) {
	at := assert.New(t)

	t.Run("success", func(t *testing.T) {
		SetupExecLookPath()
		defer TeardownExecLookPath()

		bin, err := ExecLookPath("test")

		at.Nil(err)
		at.Equal("test", bin)
	})

	t.Run("error", func(t *testing.T) {
		SetupExecLookPathError()
		defer TeardownExecLookPath()

		bin, err := ExecLookPath("test")

		at.Equal(ErrLookPath, err)
		at.Equal("", bin)
	})
}

func TestStdout(t *testing.T) {
	at := assert.New(t)

	RedirectStdout()

	_, _ = fmt.Fprint(Stdout, "stdout")

	output := DumpStdout()

	at.Equal("stdout", output)
}

func TestStderr(t *testing.T) {
	at := assert.New(t)

	RedirectStderr()

	_, _ = fmt.Fprint(Stderr, "stderr")

	output := DumpStderr()

	at.Equal("stderr", output)
}

func TestEnvs(t *testing.T) {
	at := assert.New(t)

	key1 := "DAWN_DECK"
	key2 := "DAWN_DECK2"
	oldValue := "1"
	newValue := "2"

	at.Nil(os.Setenv(key1, oldValue))

	SetupEnvs(Envs{key1: newValue, key2: newValue})

	at.Equal(newValue, os.Getenv(key1))
	at.Equal(newValue, os.Getenv(key2))

	TeardownEnvs()

	at.Equal(oldValue, os.Getenv(key1))
	at.Equal("", os.Getenv(key2))
}

func TestRunCobraCmd(t *testing.T) {
	at := assert.New(t)

	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("%v", args)
		},
	}

	out, err := RunCobraCmd(cmd, "cobra")
	at.Nil(err)
	at.Equal("[cobra]", out)
}
