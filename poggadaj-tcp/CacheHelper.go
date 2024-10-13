package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"poggadaj-tcp/universal"
)

func GetCacheConn() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", os.Getenv("CACHE_ADDRESS")),
		Password: "",
		DB:       0,
	})
}

func SetUserStatus(statusChange universal.StatusChangeMsg) {
	// Marshal the status change
	payload, err2 := json.Marshal(statusChange)
	if err2 != nil {
		Logger.Errorf("Failed to marshal status: %s", err2)
	}

	// Set user's status in cache
	err := CacheConn.Set(
		context.Background(),
		fmt.Sprintf("ggstatus:%d", statusChange.UIN),
		payload,
		0).Err()

	if err != nil {
		Logger.Errorf("Failed to set user status: %s", err)
	}

	// Publish a status change announcement
	err = CacheConn.Publish(context.Background(), "ggstatus", payload).Err()
	if err != nil {
		Logger.Errorf("Failed to publish status: %s", err)
	}
}

func GetStatusChannel() *redis.PubSub {
	return CacheConn.Subscribe(context.Background(), "ggstatus")
}

func RecvStatusChannel(pubsub *redis.PubSub) universal.StatusChangeMsg {
	statusChange := universal.StatusChangeMsg{}
	msg, err := pubsub.ReceiveMessage(context.Background())

	if err != nil {
		Logger.Errorf("Failed to receive status change: %s", err)
	}

	err = json.Unmarshal([]byte(msg.Payload), &statusChange)
	if err != nil {
		Logger.Errorf("Failed to unmarshal status change: %s", err)
	}

	return statusChange
}

func PublishMessageChannel(sender uint32, msg Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		Logger.Errorf("Failed to marshal message: %s", err)
		return err
	}

	err = CacheConn.Publish(context.Background(), fmt.Sprintf("ggmsg:%d", sender), payload).Err()
	if err != nil {
		Logger.Errorf("Failed to send message: %s", err)
	}

	Logger.Debugf("Message sent over pubsub: %s", payload)

	return err
}

func GetMessageChannel(uin uint32) *redis.PubSub {
	return CacheConn.Subscribe(context.Background(), fmt.Sprintf("ggmsg:%d", uin))
}

func RecvMessageChannel(pubsub *redis.PubSub) Message {
	message := Message{}
	msg, err := pubsub.ReceiveMessage(context.Background())

	if err != nil {
		Logger.Errorf("Failed to receive message: %s", err)
	}

	err = json.Unmarshal([]byte(msg.Payload), &message)
	if err != nil {
		Logger.Errorf("Failed to unmarshal message: %s", err)
	}

	Logger.Debugf("Message received over pubsub: %s", msg.Payload)

	return message
}

func FetchUserStatus(uin uint32) universal.StatusChangeMsg {
	statusFinal := universal.StatusChangeMsg{
		UIN:    uin,
		Status: universal.GG_STATUS_NOT_AVAIL,
	}

	status, err := CacheConn.Get(context.Background(), fmt.Sprintf("ggstatus:%d", uin)).Result()
	if err != nil {
		Logger.Errorf("Failed to fetch user status: %s", err)
		return statusFinal
	}

	err2 := json.Unmarshal([]byte(status), &statusFinal)
	if err2 != nil {
		Logger.Errorf("Failed to deserialize user status: %s", err)
		return statusFinal
	}
	return statusFinal
}
