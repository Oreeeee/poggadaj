package main

import "time"

type RegisterResponse struct {
	Error string `json:"error"`
	UIN   int    `json:"uin"`
}

type UserDataResponse struct {
	UIN    int       `json:"uin"`
	Joined time.Time `json:"joined"`
}
