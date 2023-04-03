package main

import (
	"chatgpt/router"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"time"
)

func main() {
	app := fiber.New(
		fiber.Config{
			ReadTimeout: time.Second * 20,
		})
	// 加载请求记录
	app.Use(
		logger.New(
			logger.Config{
				Format:     "${time} ${status} - ${ip}:${port} - ${method} ${path} \n",
				TimeFormat: "2006-01-02 15:04:05.000",
				TimeZone:   "Asia/Shanghai",
			},
		),
	)

	app.Use(recover.New())

	// 加载路由
	router.Api(app)
	godotenv.Load(".env")
	app.Listen(":3000")
}
