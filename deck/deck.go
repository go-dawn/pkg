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
var OsExit = func(code int) { mockOsExit(code) }
var mockOsExit = os.Exit

// ExecCommand is a wrapper for exec.Command.
var ExecCommand = func(name string, arg ...string) *exec.Cmd { return mockExecCommand(name, arg...) }
var mockExecCommand = exec.Command
var expectExecCommandError bool
var expectExecCommandStderr bool

// ExecLookPath is a wrapper for exec.LookPath.
var ExecLookPath = func(file string) (string, error) { return mockExecLookPath(file) }
var mockExecLookPath = exec.LookPath

// ErrLookPath means that error occurs when calling
// ExecLookPath.
var ErrLookPath = errors.New("deck: look path error")

// Stdout is a wrapper for os.Stdout.
var Stdout = os.Stdout
var stdoutForward *os.File

// Stderr is a wrapper for os.Stderr.
var Stderr = os.Stderr
var stderrForward *os.File

// SetupCmd mocks ExecCommand. Must create one test function
// named TestHelperCommand in a package and use HandleCommand
// in it.
func SetupCmd() {
	mockExecCommand = fakeExecCommand
}

// SetupErrorCmd mocks ExecCommand and when running the returned
// command, always get an error. Must create one test function
// named TestHelperCommand in a package and use HandleCommand
// in it.
func SetupCmdError() {
	mockExecCommand = fakeExecCommand
	expectExecCommandError = true
}

// SetupStderrCmd mocks ExecCommand. Must create one test function
// named TestHelperCommand in a package and use HandleCommand
// in it. Besides, the second parameter of the handler function in
// HandleCommand will be true.
func SetupCmdStderr() {
	mockExecCommand = fakeExecCommand
	expectExecCommandStderr = true
}

// TeardownCmd restores ExecCommand to the original one.
func TeardownCmd() {
	mockExecCommand = exec.Command
	expectExecCommandError = false
	expectExecCommandStderr = false
}

const errorCommand = "deck_exec_command_need_error"

func fakeExecCommand(command string, args ...string) (cmd *exec.Cmd) {
	if expectExecCommandError {
		return exec.Command(errorCommand)
	}

	args = append([]string{"-test.run=TestHelperCommand", "--", command}, args...)
	cmd = exec.Command(os.Args[0], args...)
	cmd.Env = []string{"GO_WANT_HELPER_COMMAND=1"}

	if expectExecCommandStderr {
		cmd.Env = append(cmd.Env, "GO_WANT_HELPER_EXPECT_STDERR=1")
	}

	return cmd
}

// HandleCommand handles every command wanted help
func HandleCommand(handler func(args []string, expectStderr bool)) {
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

	handler(args, os.Getenv("GO_WANT_HELPER_EXPECT_STDERR") == "1")

	OsExit(0)
}

// SetupExecLookPath mocks ExecLookPath.
func SetupExecLookPath() {
	mockExecLookPath = func(file string) (string, error) {
		return file, nil
	}
}

// SetupExecLookPathError mocks ExecLookPath and always return an error.
func SetupExecLookPathError() {
	mockExecLookPath = func(_ string) (string, error) {
		return "", ErrLookPath
	}
}

// TeardownExecLookPath restores ExecLookPath to the original one.
func TeardownExecLookPath() {
	mockExecLookPath = exec.LookPath
}

// SetupOsExit mocks OsExit.
func SetupOsExit(override ...func(code int)) {
	fn := func(code int) {}

	if len(override) > 0 {
		fn = override[0]
	}

	mockOsExit = fn
}

// TeardownOsExit restores OsExit to the original one.
func TeardownOsExit() {
	mockOsExit = os.Exit
}

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
