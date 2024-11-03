package structs

type Message struct {
	From     uint32 `json:"from"`
	MsgClass uint32 `json:"msg_class"`
	Content  []byte `json:"content"`
}
