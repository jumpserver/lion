package session

type Message struct {
	Opcode string      `json:"opcode"`
	Body   []string    `json:"data"`
	Meta   MetaMessage `json:"meta"` // receive的信息必须携带Meta
}

type MetaMessage struct {
	UserId  string `json:"user_id"`
	User    string `json:"user"`
	Created string `json:"created"`
}
