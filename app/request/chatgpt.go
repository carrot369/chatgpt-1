package request

type BalanceRequest struct {
	Key string `validate:"required" message:"key 不能为空" label:"key"`
}

type ChatCompletionRequest struct {
	Key           string  `validate:"-"`
	Model         string  `validate:"in:gpt-4-32k-0314,gpt-4-32k,gpt-4-0314,gpt-4,gpt-3.5-turbo-0301,gpt-3.5-turbo,text-davinci-003,text-davinci-002,text-curie-001,text-babbage-001,text-ada-001,text-davinci-001,davinci-instruct-beta,davinci,curie-instruct-beta,curie,ada,babbage" message:"in:模型选择错误" label:"模型"` // 选择模型
	Prompt        string  `validate:"required" message:"消息 不能为空" label:"消息"`
	Options       Options `validate:"-"`
	Temperature   float32 `validate:"float|min:0|lte:2" message:"numeric:随机必须为数字|min:随机性必须大于0|lte:随机性必须小于等于2" label:"随机性"` // 随机性
	DisableStream bool    `validate:"-"`
	SystemMessage string  `validate:"-"`
}

type Options struct {
	ParentMessageId string `validate:"-" json:"parentMessageId"` // 父消息ID
}
