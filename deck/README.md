# Deck
Make testing easier and under control and every package more reliable!

## Usages
### os.Exit
Use `var osExit = deck.OsExit` to replace `os.Exit`.

```go
import (
	"github.com/go-dawn/pkg/deck"
	"github.com/stretchr/testify/assert"
)

var osExit = deck.OsExit

func TestSomeFunction(t *testing.T) {
	var count int32
	deck.SetupOsExit(func(code int) {
		atomic.AddInt32(&count, int32(code))
	})
	defer deck.TeardownOsExit()

	SomeFunction()

	assert.Equal(t, int32(100), atomic.LoadInt32(&count))
}

func SomeFunction() {
	osExit(100)
}
```

### exec.Command
Use `var execCommand = deck.ExecCommand` to replace `exec.Command`. And the `TestHelperCommand` test function must be added in one of the test file in the package. You must use `deck.HandleCommand` function and determine the output of the executed command.

```go
import (
	"github.com/go-dawn/pkg/deck"
	"github.com/stretchr/testify/assert"
)

var execCommand = deck.ExecCommand

func TestSomeFunction(t *testing.T) {
	at := assert.New(t)

	t.Run("success", func(t *testing.T) {
		deck.SetupCmd()
		defer deck.TeardownCmd()

		b, err := SomeFunction()
		at.Nil(err)
		at.Equal("[test function]", string(b))
	})

	t.Run("error", func(t *testing.T) {
		deck.SetupCmdError()
		defer deck.TeardownCmd()

		b, err := SomeFunction()
		at.NotNil(err)
		at.Equal("", string(b))
	})

	t.Run("stderr", func(t *testing.T) {
		deck.SetupCmdStderr()
		defer deck.TeardownCmd()

		b, err := SomeFunction()
		at.NotNil(err)
		at.Equal("[test function]", string(b))
	})
}

func TestHelperCommand(t *testing.T) {
	deck.HandleCommand(func(args []string, expectStderr bool) {
		if expectStderr {
			_, _ = fmt.Fprintf(os.Stderr, "%v", args)
			os.Exit(1)
		}
		_, _ = fmt.Fprintf(os.Stdout, "%v", args)
	})
}

func SomeFunction() ([]byte, error) {
	cmd := execCommand("test", "function")
	return cmd.CombinedOutput()
}
```

### exec.LookPath
Use `var execLookPath = deck.ExecLookPath` to replace `exec.LookPath`.

```go
import (
	"github.com/go-dawn/pkg/deck"
	"github.com/stretchr/testify/assert"
)

var execLookPath = deck.ExecLookPath

func TestSomeFunction(t *testing.T) {
	at := assert.New(t)

	t.Run("success", func(t *testing.T) {
		SetupExecLookPath()
		defer TeardownExecLookPath()

		bin, err := SomeFunction()

		at.Nil(err)
		at.Equal("test", bin)
	})

	t.Run("error", func(t *testing.T) {
		SetupExecLookPathError()
		defer TeardownExecLookPath()

		bin, err := SomeFunction()

		at.Equal(ErrLookPath, err)
		at.Equal("", bin)
	})
}

func SomeFunction() (string, error) {
	return execLookPath("test")
}
```

### os.Stdout
Use `var stdout = deck.Stdout` to replace `os.Stdout`.

```go
import (
	"github.com/go-dawn/pkg/deck"
	"github.com/stretchr/testify/assert"
)

var stdout = deck.Stdout

func TestSomeFunction(t *testing.T) {
	at := assert.New(t)

	deck.RedirectStdout()

	SomeFunction()

	output := deck.DumpStdout()

	at.Equal("stdout", output)
}

func SomeFunction() {
	_, _ = fmt.Fprint(stdout, "stdout")
}
```

### os.Stderr
Use `var stderr = deck.Stderr` to replace `os.Stderr`.

```go
import (
	"github.com/go-dawn/pkg/deck"
	"github.com/stretchr/testify/assert"
)

var stderr = deck.Stderr

func TestSomeFunction(t *testing.T) {
	at := assert.New(t)

	deck.RedirectStderr()

	SomeFunction()

	output := deck.DumpStderr()

	at.Equal("stderr", output)
}

func SomeFunction() {
	_, _ = fmt.Fprint(stderr, "stderr")
}
```

### os.Environment
We can override or set environment and restore them back during testing.

```go
import (
	"github.com/go-dawn/pkg/deck"
	"github.com/stretchr/testify/assert"
)

func TestSomeFunction(t *testing.T) {
	at := assert.New(t)

	key1 := "DAWN_DECK"
	key2 := "DAWN_DECK2"
	oldValue := "1"
	newValue := "2"

	at.Nil(os.Setenv(key1, oldValue))

	deck.SetupEnvs(deck.Envs{key1: newValue, key2: newValue})

	v1, v2 := SomeFunction(key1, key2)

	at.Equal(newValue, v1)
	at.Equal(newValue, v2)

	deck.TeardownEnvs()

	at.Equal(oldValue, os.Getenv(key1))
	at.Equal("", os.Getenv(key2))
}

func SomeFunction(k1, k2 string) (string, string){
	return os.Getenv(k1), os.Getenv(k2)
}
```

### cobra.Command
Use `RunCobraCmd` to test a cobra command.

```go
import (
	"github.com/go-dawn/pkg/deck"
	"github.com/stretchr/testify/assert"
)

func TestRunCobraCmd(t *testing.T) {
	at := assert.New(t)

	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("%v", args)
		},
	}

	out, err := deck.RunCobraCmd(cmd, "cobra")

	at.Nil(err)
	at.Equal("[cobra]", out)
}
```

### httptest
Use `SetupServer` to get `*fiber.App` instance as `app` and `*httptest.Expect` instance as `e`. And then register routes by `app`. Next make request by `e` and finally do assertion with several helper functions. 

```go
import (
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-dawn/dawn/fiberx"
	"github.com/go-dawn/pkg/deck"
	"github.com/gofiber/fiber/v2"
)

func Test_Fiber_Routes(t *testing.T) {
	app, e := deck.SetupServer(t)

	app.Get("/", func(c *fiber.Ctx) error {
		return fiberx.Message(c, "test")
	})

	resp := e.GET("/").Expect()

	deck.AssertRespStatus(resp, fiber.StatusOK)
	deck.AssertRespCode(resp, 200)
	deck.AssertRespMsg(resp, "test")
	deck.AssertRespMsgContains(resp, "es")

	data := deck.Expected{
		Status:  500,
		Code:    500,
		Msg:     "error",
		Contain: false,
	}

	app.Get("/data", func(c *fiber.Ctx) error {
		return fiberx.Data(c, data)
	})

	resp2 := e.GET("/data").Expect()

	deck.AssertRespData(resp2, data)
	deck.AssertRespDataCheck(resp2, func(v *httpexpect.Value) {
		obj := v.Object()
		obj.ValueEqual("Status", 500)
		obj.ValueEqual("Code", 500)
		obj.ValueEqual("Msg", "error")
		obj.ValueEqual("Contain", false)
	})
}
```
