package app

import (
	"github.com/imroc/req/v3"
	"os"
	"time"
)

var ChatGptBaseUrl = "https://api.openai.com/"

var BaseClient = req.C().SetBaseURL(ChatGptBaseUrl).SetTimeout(10 * time.Second)

func GetClient() *req.Client {
	if os.Getenv("DEV") == "true" {
		return BaseClient.SetProxyURL("http://127.0.0.1:7890").DevMode()
	}
	return BaseClient
}

type BalanceResp struct {
	Object         string  `json:"object"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
	TotalAvailable float64 `json:"total_available"`
	Grants         struct {
		Object string `json:"object"`
		Data   []struct {
			Object      string  `json:"object"`
			Id          string  `json:"id"`
			GrantAmount float64 `json:"grant_amount"`
			UsedAmount  float64 `json:"used_amount"`
			EffectiveAt float64 `json:"effective_at"`
			ExpiresAt   float64 `json:"expires_at"`
		} `json:"data"`
	} `json:"grants"`
}
