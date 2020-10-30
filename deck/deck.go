package deck

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// OsExit is a wrapper for os.Exit.
var OsExit = os.Exit

// ExecCommand is a wrapper for exec.Command.
var ExecCommand = exec.Command

// ExecLookPath is a wrapper for exec.LookPath.
var ExecLookPath = exec.LookPath

// ErrLookPath means that error occurs when calling
// ExecLookPath.
var ErrLookPath = errors.New("deck: look path error")

// NeedErr determines whether command or function returns error
var NeedErr = struct{}{}

// Stdout is a wrapper for os.Stdout.
var Stdout = os.Stdout

// Stderr is a wrapper for os.Stderr.
var Stderr = os.Stderr

var needError bool

// SetupCmd mocks ExecCommand. Should create one test function
// named TestHelperCommand in a package and use HandleCommand
// in it. See it for more detail.
func SetupCmd(flag ...struct{}) {
	ExecCommand = fakeExecCommand
	if len(flag) > 0 {
		needError = true
	}
}

// TeardownCmd restores ExecCommand to the original one.
func TeardownCmd() {
	ExecCommand = exec.Command
	needError = false
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperCommand", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_COMMAND=1"}
	if needError {
		cmd.Env = append(cmd.Env, "GO_WANT_HELPER_NEED_ERR=1")
	}
	return cmd
}

// HandleCommand handles every command
func HandleCommand(handler func(args []string, needErr bool)) {
	if os.Getenv("GO_WANT_HELPER_COMMAND") != "1" {
		return
	}
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	handler(args, os.Getenv("GO_WANT_HELPER_NEED_ERR") == "1")

	OsExit(0)
}

// SetupExecLookPath mocks ExecLookPath. Pass NeedErr if you want an error.
func SetupExecLookPath(flag ...struct{}) {
	ExecLookPath = func(file string) (s string, err error) {
		s = file

		if len(flag) > 0 {
			err = ErrLookPath
		}
		return
	}
}

// TeardownExecLookPath restores ExecLookPath to the original one.
func TeardownExecLookPath() {
	ExecLookPath = exec.LookPath
}

// SetupOsExit mocks OsExit.
func SetupOsExit(fn ...func(code int)) {
	f := func(code int) {}

	if len(fn) > 0 {
		f = fn[0]
	}

	OsExit = f
}

// TeardownOsExit restores OsExit to the original one.
func TeardownOsExit() {
	OsExit = os.Exit
}

var stdoutForward *os.File

// RedirectStdout mocks Stdout.
func RedirectStdout() {
	stdoutForward, Stdout, _ = os.Pipe()
}

// DumpStdout dumps output from Stdout and restores it to the original one.
func DumpStdout() string {
	_ = Stdout.Close()

	b, _ := ioutil.ReadAll(stdoutForward)

	Stdout = os.Stdout

	return string(b)
}

var stderrForward *os.File

// RedirectStderr mocks Stderr.
func RedirectStderr() {
	stderrForward, Stderr, _ = os.Pipe()
}

// DumpStderr dumps output from Stderr and restores it to the original one.
func DumpStderr() string {
	_ = Stderr.Close()

	b, _ := ioutil.ReadAll(stderrForward)

	Stderr = os.Stderr

	return string(b)
}

// Envs is used for override or set env
type Envs map[string]string

var oldEnvs Envs

const nonExistEnv = "deck: non exist"

// SetupEnvs can override or set envs
func SetupEnvs(envs Envs) {
	oldEnvs = make(Envs, len(envs))

	for k, v := range envs {
		if val, ok := os.LookupEnv(k); ok {
			oldEnvs[k] = val
		} else {
			oldEnvs[k] = nonExistEnv
		}
		_ = os.Setenv(k, v)
	}
}

// TeardownEnvs restores envs to original ones
func TeardownEnvs() {
	for k, v := range oldEnvs {
		_ = os.Unsetenv(k)
		if v != nonExistEnv {
			_ = os.Setenv(k, v)
		}
	}
}

// RunCobraCmd executes a cobra command and get output and error
func RunCobraCmd(cmd *cobra.Command, args ...string) (string, error) {
	var b bytes.Buffer

	cmd.ResetCommands()
	cmd.SetErr(&b)
	cmd.SetOut(&b)
	cmd.SetArgs(args)

	err := cmd.Execute()

	return b.String(), err
}
