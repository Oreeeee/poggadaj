package universal

type StatusChangeMsg struct {
	UIN    uint32 `json:"uin"`
	Status uint32 `json:"status"`
}
