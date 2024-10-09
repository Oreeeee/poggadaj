package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

func GetCacheConn() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "cache:6379",
		Password: "",
		DB:       0,
	})
}

func SetUserStatus(uin uint32, status uint32) error {
	return CacheConn.Set(
		context.Background(),
		fmt.Sprintf("ggstatus:%d", uin),
		status,
		0).Err()
}
