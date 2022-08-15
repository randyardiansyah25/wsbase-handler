package wsbase

const (
	TypeBroadcast = iota
	TypePrivate
)

type Message struct {
	Type     int         `json:"type"`
	SenderId string      `json:"sender_id"`
	To       []string    `json:"to"`
	Action   string      `json:"action"`
	Title    string      `json:"title"`
	Body     interface{} `json:"body"`
}
