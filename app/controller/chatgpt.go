package controller

import (
	"chatgpt/app"
	"chatgpt/app/request"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gookit/validate"
	"log"
)

// GetBalance 查询账户余额
func GetBalance() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req request.BalanceRequest
		if err := c.QueryParser(&req); err != nil {
			return app.Error(c, "解析参数错误："+err.Error())
		}

		v := validate.New(req)
		if !v.Validate() {
			return app.Error(c, v.Errors.One())
		}

		// 查询余额
		var result app.BalanceResp
		if _, err := app.BaseClient.DevMode().R().
			SetContentType("application/json").
			SetSuccessResult(&result).
			SetBearerAuthToken(req.Key).
			Get("dashboard/billing/credit_grants"); err != nil {
			log.Println(fmt.Sprintf("查询余额失败：%s", err.Error()))
			return app.Error(c, "查询余额失败")
		}

		return app.Success(c, result)
	}
}
