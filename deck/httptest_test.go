package deck

import (
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/go-dawn/dawn/fiberx"
	"github.com/gofiber/fiber/v2"
)

func Test_Httptest_AssertResp(t *testing.T) {
	app, e := SetupServer(t)

	app.Get("/", func(c *fiber.Ctx) error {
		return fiberx.Message(c, "test")
	})

	resp := e.GET("/").Expect()

	AssertRespStatus(resp, fiber.StatusOK)
	AssertRespCode(resp, 200)
	AssertRespMsg(resp, "test")
	AssertRespMsgContains(resp, "es")

	data := Expected{
		Status:  500,
		Code:    500,
		Msg:     "error",
		Contain: false,
	}

	app.Get("/data", func(c *fiber.Ctx) error {
		return fiberx.Data(c, data)
	})

	resp2 := e.GET("/data").Expect()

	AssertRespData(resp2, data)
	AssertRespDataCheck(resp2, func(v *httpexpect.Value) {
		obj := v.Object()
		obj.ValueEqual("Status", 500)
		obj.ValueEqual("Code", 500)
		obj.ValueEqual("Msg", "error")
		obj.ValueEqual("Contain", false)
	})
}
