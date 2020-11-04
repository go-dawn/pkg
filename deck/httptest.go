package deck

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-dawn/dawn/fiberx"
	"github.com/gofiber/fiber/v2"
)

// Expected holds everything need to assert
type Expected struct {
	// Status is response status code
	Status int
	// Code is business code
	Code int
	// Msg is response message
	Msg string
	// Contain indicates
	Contain bool
	// Data is response data
	Data interface{}
	// DataChecker is a function which gives you ability
	// to make assertion with data
	DataChecker func(data *httpexpect.Value) `json:"-"`
}

// SetupServer gets fiber.App and httpexpect.Expect instance
func SetupServer(t *testing.T) (*fiber.App, *httpexpect.Expect) {
	app := fiber.New(fiber.Config{ErrorHandler: fiberx.ErrHandler})

	return app, httpexpect.WithConfig(httpexpect.Config{
		// Pass requests directly to FastHTTPHandler.
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(app.Handler()),
			Jar:       httpexpect.NewJar(),
		},
		// Report errors using testify.
		Reporter: httpexpect.NewAssertReporter(t),
	})
}

// AssertRespStatus asserts response with an expected status code
func AssertRespStatus(resp *httpexpect.Response, status int) {
	AssertResp(resp, Expected{Status: status})
}

// AssertRespCode asserts response with an expected business code
func AssertRespCode(resp *httpexpect.Response, code int) {
	AssertResp(resp, Expected{Code: code})
}

// AssertRespMsg asserts response with an expected message
func AssertRespMsg(resp *httpexpect.Response, msg string) {
	AssertResp(resp, Expected{Msg: msg})
}

// AssertRespMsgContains asserts response contains an expected message
func AssertRespMsgContains(resp *httpexpect.Response, msg string) {
	AssertResp(resp, Expected{Msg: msg, Contain: true})
}

// AssertRespData asserts response with an expected data
func AssertRespData(resp *httpexpect.Response, data interface{}) {
	AssertResp(resp, Expected{Data: data})
}

// AssertRespDataCheck asserts response with an expected data checker
func AssertRespDataCheck(resp *httpexpect.Response, dataChecker func(value *httpexpect.Value)) {
	AssertResp(resp, Expected{DataChecker: dataChecker})
}

// AssertResp asserts response with an Expected instance
func AssertResp(resp *httpexpect.Response, r Expected) {
	if r.Status != 0 {
		resp.Status(r.Status)
	}

	obj := resp.JSON().Object()

	if r.Code != 0 {
		obj.ValueEqual("code", r.Code)
	}

	if r.Msg != "" {
		if r.Contain {
			obj.Value("message").String().Contains(r.Msg)
		} else {
			obj.ValueEqual("message", r.Msg)
		}
	}

	if r.Data != nil {
		obj.ValueEqual("data", r.Data)
	}

	if r.DataChecker != nil {
		r.DataChecker(obj.Value("data"))
	}
}
