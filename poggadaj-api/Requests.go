package main

type RegisterRequest struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	GGAncientPassword string `json:"gg_ancient_password"`
	GG32Password      string `json:"gg32_password"`
	GGSHA1Password    string `json:"gg_sha1_password"`
}
