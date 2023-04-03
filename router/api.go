package router

import (
	"chatgpt/app/controller"
	"github.com/gofiber/fiber/v2"
)

func Api(app *fiber.App) {

	api := app.Group("/api")
	api.Get("/get-balance", controller.GetBalance())             // 查询余额
	api.Post("/chat-process", controller.CreateChatCompletion()) // 发送聊天
	api.Post("/session", controller.CreateSession())             // 创建会话
}
