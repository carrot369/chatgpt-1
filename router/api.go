package router

import (
	"chatgpt/app/controller"
	"github.com/gofiber/fiber/v2"
)

func Api(app *fiber.App) {

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// chatgpt
	chatgpt := v1.Group("/chatgpt")
	chatgpt.Get("/get_balance", controller.GetBalance())

}
