package app

import (
	"github.com/gofiber/fiber/v2"
)

// Response 统一返回
func Response(c *fiber.Ctx, code int, data interface{}, msg string) error {
	return c.Status(code).JSON(fiber.Map{
		"code": code,
		"data": data,
		"msg":  msg,
	})
}

// Success 成功返回
func Success(c *fiber.Ctx, data interface{}) error {
	return Response(c, fiber.StatusOK, data, "success")
}

// Error 失败返回
func Error(c *fiber.Ctx, msg string) error {
	return Response(c, fiber.StatusInternalServerError, nil, msg)
}
