package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"poggadaj-tcp/universal"
	"strconv"
)

func GetCacheConn() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", os.Getenv("CACHE_ADDRESS")),
		Password: "",
		DB:       0,
	})
}

func SetUserStatus(uin uint32, status uint32) {
	// Set user's status in cache
	err := CacheConn.Set(
		context.Background(),
		fmt.Sprintf("ggstatus:%d", uin),
		status,
		0).Err()

	if err != nil {
		fmt.Println("Failed to set user status:", err)
	}

	// Publish a status change announcement
	payload, err2 := json.Marshal(universal.StatusChangeMsg{uin, status})
	if err2 != nil {
		fmt.Println("Failed to marshal status:", err2)
	}

	err = CacheConn.Publish(context.Background(), "ggstatus", payload).Err()
	if err != nil {
		fmt.Println("Failed to publish status:", err)
	}
}

func GetStatusChannel() *redis.PubSub {
	return CacheConn.Subscribe(context.Background(), "ggstatus")
}

func RecvStatusChannel(pubsub *redis.PubSub) universal.StatusChangeMsg {
	statusChange := universal.StatusChangeMsg{}
	msg, err := pubsub.ReceiveMessage(context.Background())

	if err != nil {
		fmt.Println("Failed to receive status change:", err)
	}

	err = json.Unmarshal([]byte(msg.Payload), &statusChange)
	if err != nil {
		fmt.Println("Failed to unmarshal status change:", err)
	}

	return statusChange
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

func FetchUserStatus(uin uint32) uint32 {
	status, err := CacheConn.Get(context.Background(), fmt.Sprintf("ggstatus:%d", uin)).Result()
	if err != nil {
		fmt.Println("Failed to fetch user status:", err)
	}

	statusInt, err2 := strconv.Atoi(status)
	if err2 != nil {
		fmt.Println("Failed to fetch user status:", err)
	}
	return uint32(statusInt)
}
