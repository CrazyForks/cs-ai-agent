package eventbus

type UserCreated struct {
	UserID int64
	Name   string
}

type OrderCreated struct {
	OrderID int64
	UserID  int64
	Amount  int64
}
