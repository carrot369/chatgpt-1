package response

import "github.com/sashabaranov/go-openai"

type BalanceResponse struct {
	Key     string  `json:"key"`
	Total   float64 `json:"total"`
	Balance float64 `json:"balance"`
	Used    float64 `json:"used"`
}

type ChatCompletionResponse struct {
	Role            string                         `json:"role"`
	Id              string                         `json:"id"`
	ParentMessageId string                         `json:"parentMessageId"`
	Text            string                         `json:"text"`
	Detail          *openai.ChatCompletionResponse `json:"detail"`
}

type ChatCompletionStreamResponse struct {
	Role            string                               `json:"role"`
	Id              string                               `json:"id"`
	ParentMessageId string                               `json:"parentMessageId"`
	Text            string                               `json:"text"`
	Detail          *openai.ChatCompletionStreamResponse `json:"detail"`
}

type SessionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Auth  bool   `json:"auth"`
		Model string `json:"model"`
	} `json:"data"`
}

type ErrorStream struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Status  string      `json:"status"`
}
