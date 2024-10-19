package main

type RegisterResponse struct {
	Error string `json:"error"`
	UIN   int    `json:"uin"`
}
