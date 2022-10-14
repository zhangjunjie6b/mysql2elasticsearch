package consume

type ConsumeInterface interface {
	Handle(data interface{}) error
}
