package controller

import (
	"bufio"
	"chatgpt/app"
	"chatgpt/app/request"
	"chatgpt/app/response"
	"chatgpt/app/service"
	"chatgpt/database"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gookit/validate"
	"github.com/sashabaranov/go-openai"
	"io"
	"log"
	"os"
	"strings"
	"time"
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
		result, err := service.GetBalance(req.Key)
		if err != nil {
			return app.Error(c, err.Error())
		}

		return app.Success(c, response.BalanceResponse{
			Total:   result.TotalGranted,
			Used:    result.TotalUsed,
			Balance: result.TotalAvailable,
		})
	}
}

// CreateChatCompletion 发送聊天
func CreateChatCompletion() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req request.ChatCompletionRequest
		if err := c.BodyParser(&req); err != nil {
			return app.Error(c, "解析参数错误："+err.Error())
		}

		v := validate.New(req)
		if !v.Validate() {
			return app.Error(c, v.Errors.One())
		}

		chatGPT := service.ChatGPTService{
			Key:   req.Key,
			Proxy: os.Getenv("PROXY"),
			ChatCompletionRequest: openai.ChatCompletionRequest{
				Model:       req.Model,
				Temperature: req.Temperature,
				Stream:      true,
			},
		}

		if req.DisableStream {
			chatGPT.ChatCompletionRequest.Stream = false
		}

		if err := chatGPT.ContextHandler(req); err != nil {
			return app.Error(c, ErrorHandler(req.Key, err).Error())
		}

		// 创建聊天
		if err := chatGPT.CreateChatCompletion(); err != nil {
			return app.Error(c, ErrorHandler(req.Key, err).Error())
		}

		if req.DisableStream {
			var text string
			for _, message := range chatGPT.Response.ChatCompletionResponse.Choices {
				text += message.Message.Content
			}
			database.GPTCache.Add(chatGPT.Response.ChatCompletionResponse.ID, 1*time.Hour, database.GPTCacheItem{
				NowID:    chatGPT.Response.ChatCompletionResponse.ID,
				Prompt:   req.Prompt,
				ParentID: req.Options.ParentMessageId,
				Answer:   text,
			})
			return app.Success(c, response.ChatCompletionResponse{
				Role:   openai.ChatMessageRoleAssistant,
				Id:     chatGPT.Response.ChatCompletionResponse.ID,
				Text:   text,
				Detail: chatGPT.Response.ChatCompletionResponse,
			})
		}
		var gptResult response.ChatCompletionStreamResponse
		// fiber 返回steam
		c.Response().Header.SetContentType("application/octet-stream")
		c.Response().Header.Set("Transfer-Encoding", "chunked")
		c.Response().Header.Set("Keep-Alive", "timeout=4")
		c.Response().Header.Set("Proxy-Connection", "keep-alive")
		c.Response().Header.Set("connection", "keep-alive")
		c.Response().Header.Set("Access-Control-Allow-Methods", "*")
		c.Response().Header.Set("Access-Control-Allow-Origin", "*")
		c.Response().Header.Set("Access-Control-Allow-Headers", "authorization, Content-Type")
		c.Response().SetBodyStreamWriter(func(w *bufio.Writer) {
			for {
				if chatGPT.Response.Stream != nil {
					r, err := chatGPT.Response.Stream.Recv()
					if errors.Is(err, io.EOF) {
						if gptResult.Id != "" && !database.GPTCache.Exists(gptResult.Id) {
							database.GPTCache.Add(gptResult.Id, 1*time.Hour, database.GPTCacheItem{
								NowID:    gptResult.Id,
								ParentID: req.Options.ParentMessageId,
								Prompt:   req.Prompt,
								Answer:   gptResult.Text,
							})
						}
						return
					}
					if err != nil {
						e := response.ErrorStream{
							Message: ErrorHandler(req.Key, err).Error(),
							Data:    nil,
							Status:  "Fail",
						}

						marshal, _ := json.Marshal(e)
						if _, err := fmt.Fprintf(w, "%s\n", marshal); err != nil {
							fmt.Println(err)
							return
						}
						_ = w.Flush()
						return
					}

					if len(r.Choices) == 0 {
						fmt.Println("response.Choices is empty")
						_ = w.Flush()
						return
					}
					gptResult.Detail = &r
					gptResult.Id = r.ID
					gptResult.Role = openai.ChatMessageRoleAssistant
					gptResult.Text += r.Choices[0].Delta.Content

					marshal, _ := json.Marshal(gptResult)
					if _, err := fmt.Fprintf(w, "%s\n", marshal); err != nil {
						fmt.Println(err)
						return
					}

					if err := w.Flush(); err != nil {
						fmt.Println(err)
						return
					}
				}
			}
		})
		return nil
	}
}

func ErrorHandler(key string, err error) error {
	log.Println(err)
	// 无效key
	if strings.Contains(err.Error(), "Incorrect API key provided") {
		return errors.New("无效key，请检查key是否正确")

	}
	// 余额不足
	if strings.Contains(err.Error(), "You exceeded your current quota") {
		if key == "" {
			// 更换key
			if errs := service.ChangeKey(); errs != nil {
				return errors.New("更换key失败，请联系管理员")
			}
			return errors.New("请重试")

		}
		return errors.New("余额不足，请充值, 或更换key")
	}
	if strings.Contains(err.Error(), "You didn't provide an API key") {
		return errors.New("未提供key，请提供key")
	}
	if strings.Contains(err.Error(), "Rate limit reached for") {
		return errors.New("当前请求次数过多，请稍后再试即可")
	}
	// 未知错误
	if err != nil {
		return errors.New("未知错误，请联系管理员")
	}
	return err
}

func CreateSession() fiber.Handler {
	return func(c *fiber.Ctx) error {

		return c.JSON(response.SessionResponse{
			Status: "Success",
			Data: struct {
				Auth  bool   `json:"auth"`
				Model string `json:"model"`
			}(struct {
				Auth  bool
				Model string
			}{Auth: false, Model: "ChatGPTAPI"}),
		})
	}
}
