package mode

type Jobs struct {
	ID uint `gorm:"primaryKey"`
	Queue string
	Payload   Payloads `gorm:"serializer:json"`
	Del string
	Attempts int
	LastError string
}

type Payloads struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
	EsIndexName string `json:"name"`
}
