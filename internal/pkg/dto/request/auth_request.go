package request

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type WxWorkExchangeRequest struct {
	Ticket string `json:"ticket"`
}
