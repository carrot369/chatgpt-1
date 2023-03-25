package request

type BalanceRequest struct {
	Key string `validate:"required" message:"key 不能为空" label:"key"`
}
