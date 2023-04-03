package service

import (
	"bufio"
	"bytes"
	"chatgpt/app"
	"chatgpt/app/request"
	"chatgpt/database"
	"context"
	"fmt"
	aiToken "github.com/andreyvit/openai"
	"github.com/sashabaranov/go-openai"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

const (
	defaultModel       = openai.GPT3Dot5Turbo0301
	defaultTemperature = 0.8
	defaultN           = 1
)

var ModelsMaxToken = map[string]int{
	openai.GPT3Dot5Turbo:       4096,
	openai.GPT3Dot5Turbo0301:   4096,
	openai.GPT4:                8192,
	openai.GPT40314:            8192,
	openai.GPT432K:             32768,
	openai.GPT432K0314:         32768,
	openai.CodexCodeDavinci002: 8001,
	openai.CodexCodeDavinci001: 8001,
	openai.CodexCodeCushman001: 2048,
	openai.GPT3TextDavinci003:  4097,
	openai.GPT3TextDavinci002:  4097,
	openai.GPT3TextCurie001:    2049,
	openai.GPT3TextBabbage001:  2049,
	openai.GPT3TextAda001:      2049,
	openai.GPT3Davinci:         2049,
	openai.GPT3Curie:           2049,
	openai.GPT3Ada:             2049,
	openai.GPT3Babbage:         2049,
}

type ChatGPTService struct {
	Key                   string                       `json:"key"`
	Proxy                 string                       `json:"proxy"` // 代理
	ChatCompletionRequest openai.ChatCompletionRequest `json:"chatCompletionRequest"`
	Response              ChatGPTResponse              `json:"response"`
}

type ChatGPTResponse struct {
	ChatCompletionResponse *openai.ChatCompletionResponse `json:"chatCompletionResponse"`
	Stream                 *openai.ChatCompletionStream
}

// GetBalance 查询账户余额
func GetBalance(key string) (app.BalanceResp, error) {
	// 查询余额
	var result app.BalanceResp
	if _, err := app.GetClient().R().
		SetContentType("application/json").
		SetSuccessResult(&result).
		SetBearerAuthToken(key).
		Get("dashboard/billing/credit_grants"); err != nil {
		log.Println(fmt.Sprintf("查询余额失败：%s", err.Error()))
		return result, fmt.Errorf("查询余额失败")
	}
	return result, nil
}

// CreateChatCompletion 创建聊天
func (c *ChatGPTService) CreateChatCompletion() error {
	if c.Key == "" {
		key, err := GetNewKey()
		if err != nil {
			return fmt.Errorf("获取key失败：%s", err.Error())
		}
		c.Key = key
	}

	config := openai.DefaultConfig(c.Key)

	// 设置代理
	if c.Proxy != "" {
		proxyUrl, err := url.Parse(c.Proxy)
		if err != nil {
			return fmt.Errorf("代理设置失败：%s", err.Error())
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		config.HTTPClient = &http.Client{
			Transport: transport,
		}
	}

	if c.ChatCompletionRequest.Model == "" {
		c.ChatCompletionRequest.Model = defaultModel
	}
	if c.ChatCompletionRequest.Temperature == 0 {
		c.ChatCompletionRequest.Temperature = defaultTemperature
	}
	if c.ChatCompletionRequest.N == 0 {
		c.ChatCompletionRequest.N = defaultN
	}

	client := openai.NewClientWithConfig(config)
	ctx := context.Background()

	if c.ChatCompletionRequest.Stream {
		stream, err := client.CreateChatCompletionStream(ctx, c.ChatCompletionRequest)
		if err != nil {
			return fmt.Errorf("创建聊天失败：%s", err.Error())
		}
		c.Response.Stream = stream
	} else {
		resp, err := client.CreateChatCompletion(ctx, c.ChatCompletionRequest)
		if err != nil {
			return fmt.Errorf("创建聊天失败：%s", err.Error())
		}
		c.Response.ChatCompletionResponse = &resp
	}
	return nil
}

// ContextHandler 上下文处理
func (c *ChatGPTService) ContextHandler(req request.ChatCompletionRequest) error {
	var str strings.Builder
	model := c.ChatCompletionRequest.Model
	if model == "" {
		model = defaultModel
	}

	maxToken, ok := ModelsMaxToken[model]
	if !ok {
		return fmt.Errorf("不支持的模型：" + c.ChatCompletionRequest.Model)
	}
	// 本次请求
	c.ChatCompletionRequest.Messages = append(c.ChatCompletionRequest.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Prompt,
	})
	str.WriteString(req.Prompt)

	// 历史请求
	id := req.Options.ParentMessageId
	i := 0
	for {
		if id == "" {
			break
		}

		value, _ := database.GPTCache.Value(id)
		if value == nil {
			break
		}

		gptCacheItem := value.Data().(database.GPTCacheItem)

		str.WriteString(gptCacheItem.Prompt)
		str.WriteString(gptCacheItem.Answer)

		len := aiToken.TokenCount(str.String(), model)
		if len >= maxToken {
			break
		}
		fmt.Println("len: ", len, "maxToken: ", maxToken, "i: ", i)

		c.ChatCompletionRequest.Messages = append(c.ChatCompletionRequest.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: gptCacheItem.Prompt,
		},
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: gptCacheItem.Answer,
			},
		)

		id = gptCacheItem.ParentID
		i += 1
	}

	if req.SystemMessage != "" {
		c.ChatCompletionRequest.Messages = append(c.ChatCompletionRequest.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemMessage,
		})
	}

	// 倒序
	if len(c.ChatCompletionRequest.Messages) > 1 {
		for i, j := 0, len(c.ChatCompletionRequest.Messages)-1; i < j; i, j = i+1, j-1 {
			c.ChatCompletionRequest.Messages[i], c.ChatCompletionRequest.Messages[j] = c.ChatCompletionRequest.Messages[j], c.ChatCompletionRequest.Messages[i]
		}
	}
	return nil
}

// ChangeKey 更换key
func ChangeKey() error {
	key, err := GetNewKey()
	if err != nil {
		return err
	}

	// 余额不足
	log.Println("额度不足，开始更换key")
	log.Println("当前key：" + key)

	// 删除当前key
	if err := DeleteKey(key); err != nil {
		return err
	}

	// 写入不可用的key
	if err := WriteDisableKey(key); err != nil {
		return err
	}

	return nil
}

// GetNewKey 获得新的key
func GetNewKey() (string, error) {
	// 读取运行目录下的 enable.txt 的第一行字符串
	file, err := os.Open("enable.txt")
	if err != nil {
		log.Println(fmt.Sprintf("读取enable.txt失败：%s", err.Error()))
	}
	defer file.Close()

	// 取第一行
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(), nil
	}

	return "", fmt.Errorf("读取enable.txt失败: enable.txt文件为空")
}

// DeleteKey 删除key
func DeleteKey(key string) error {
	// 读取整个文件内容
	file, err := os.Open("enable.txt")
	if err != nil {
		log.Println(fmt.Sprintf("读取enable.txt失败：%s", err.Error()))
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) == 0 {
		return fmt.Errorf("enable.txt文件为空")
	}

	if lines[0] != key {
		return fmt.Errorf("enable.txt文件第一行不是当前key")
	}

	// 删除第一行
	lines = lines[1:]

	// 写入文件
	var buffer bytes.Buffer
	for _, line := range lines {
		buffer.WriteString(line + "\n")
	}

	if err := ioutil.WriteFile("enable.txt", buffer.Bytes(), 0666); err != nil {
		return fmt.Errorf("写入enable.txt失败：%s", err.Error())
	}

	return nil

}

// WriteDisableKey 写入不可用的key
func WriteDisableKey(key string) error {
	// 追加写入
	file, err := os.OpenFile("disable.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("Error opening file: %v\n", err)
	}
	defer file.Close()
	if _, err = file.WriteString(key + "\n"); err != nil {
		return fmt.Errorf("Error writing file: %v\n", err)
	}
	return nil
}

// RestartChatGPTWeb 重启chatGPT-web
func RestartChatGPTWeb() error {
	// 执行命令 supervisorctl restart chatgpt-web
	cmd := exec.Command("supervisorctl", "restart", "chatgpt-web")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error executing command: %s\n", err)

	}
	log.Println(string(output))
	return nil
}
