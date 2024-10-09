package main

import (
	"context"
	"encoding/json"
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
	err := CacheConn.Set(
		context.Background(),
		fmt.Sprintf("ggstatus:%d", uin),
		status,
		0).Err()

	if err != nil {
		fmt.Println("Failed to set user status:", err)
	}
	return err
}

func PublishMessageChannel(sender uint32, msg Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Failed to marshal message:", err)
		return err
	}

	err = CacheConn.Publish(context.Background(), fmt.Sprintf("ggmsg:%d", sender), payload).Err()
	if err != nil {
		fmt.Println("Failed to send message:", err)
	}

	fmt.Println("Message sent over pubsub:", payload)

	return err
}

func GetMessageChannel(uin uint32) *redis.PubSub {
	return CacheConn.Subscribe(context.Background(), fmt.Sprintf("ggmsg:%d", uin))
}

func RecvMessageChannel(pubsub *redis.PubSub) Message {
	message := Message{}
	msg, err := pubsub.ReceiveMessage(context.Background())

	if err != nil {
		fmt.Println("Failed to receive message:", err)
	}

	err = json.Unmarshal([]byte(msg.Payload), &message)
	if err != nil {
		fmt.Println("Failed to unmarshal message:", err)
	}

	fmt.Println("Message received over pubsub:", msg.Payload)

	return message
}
