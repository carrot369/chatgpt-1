package main

import (
	"chatgpt/router"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	// 加载请求记录
	app.Use(
		logger.New(
			logger.Config{
				Format:     "${time} - ${ip}:${port} - ${method} ${path} \n",
				TimeFormat: "2006-01-02 15:04:05.000",
				TimeZone:   "Asia/Shanghai",
			},
		),
	)

	// 加载路由
	router.Api(app)

	app.Listen(":3000")
}
