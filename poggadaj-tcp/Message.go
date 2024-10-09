package main

type Message struct {
	From    uint32 `json:"from"`
	Content []byte `json:"content"`
}
