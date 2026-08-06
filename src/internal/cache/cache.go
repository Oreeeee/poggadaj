// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"charm.land/log/v2"
	"codeberg.org/or3e/poggadaj/internal/statuses"
	"codeberg.org/or3e/poggadaj/internal/structs"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	conn   *redis.Client
	logger *log.Logger
}

func NewCache(logger *log.Logger) (*Cache, error) {
	cache := &Cache{
		conn: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:6379", os.Getenv("CACHE_ADDRESS")),
			Password: "",
			DB:       0,
		}),
		logger: logger,
	}

	return cache, nil
}

func (cache *Cache) SetUserStatus(statusChange structs.StatusChangeMsg) {
	// Marshal the status change
	payload, err2 := json.Marshal(statusChange)
	if err2 != nil {
		cache.logger.Error("Failed to marshal status", "err", err2)
	}

	// Set user's status in cache
	err := cache.conn.Set(
		context.Background(),
		fmt.Sprintf("ggstatus:%d", statusChange.UIN),
		payload,
		0).Err()

	if err != nil {
		cache.logger.Error("Failed to set user status", "err", err)
	}

	// Publish a status change announcement
	err = cache.conn.Publish(context.Background(), "ggstatus", payload).Err()
	if err != nil {
		cache.logger.Error("Failed to publish status", "err", err)
	}
}

func (cache *Cache) GetStatusChannel() *redis.PubSub {
	return cache.conn.Subscribe(context.Background(), "ggstatus")
}

func (cache *Cache) RecvStatusChannel(pubsub *redis.PubSub) structs.StatusChangeMsg {
	statusChange := structs.StatusChangeMsg{}
	msg, err := pubsub.ReceiveMessage(context.Background())

	if err != nil {
		cache.logger.Error("Failed to receive status change", "err", err)
	}

	err = json.Unmarshal([]byte(msg.Payload), &statusChange)
	if err != nil {
		cache.logger.Error("Failed to unmarshal status change", err)
	}

	return statusChange
}

func (cache *Cache) PublishMessageChannel(sender uint32, msg structs.Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		cache.logger.Error("Failed to marshal message", "err", err)
		return err
	}

	err = cache.conn.Publish(context.Background(), fmt.Sprintf("ggmsg:%d", sender), payload).Err()
	if err != nil {
		cache.logger.Error("Failed to send message", "err", err)
	}

	cache.logger.Debugf("Message sent over pubsub: %s", payload)

	return err
}

func (cache *Cache) GetMessageChannel(uin uint32) *redis.PubSub {
	return cache.conn.Subscribe(context.Background(), fmt.Sprintf("ggmsg:%d", uin))
}

func (cache *Cache) RecvMessageChannel(pubsub *redis.PubSub) structs.Message {
	message := structs.Message{}
	msg, err := pubsub.ReceiveMessage(context.Background())

	if err != nil {
		cache.logger.Error("Failed to receive message", "err", err)
	}

	err = json.Unmarshal([]byte(msg.Payload), &message)
	if err != nil {
		cache.logger.Error("Failed to unmarshal message", "err", err)
	}

	cache.logger.Debugf("Message received over pubsub: %s", msg.Payload)

	return message
}

func (cache *Cache) FetchUserStatus(uin uint32) structs.StatusChangeMsg {
	statusFinal := structs.StatusChangeMsg{
		UIN:    uin,
		Status: statuses.GG_STATUS_NOT_AVAIL,
	}

	status, err := cache.conn.Get(context.Background(), fmt.Sprintf("ggstatus:%d", uin)).Result()
	if err != nil {
		cache.logger.Error("Failed to fetch user status", "err", err)
		return statusFinal
	}

	err2 := json.Unmarshal([]byte(status), &statusFinal)
	if err2 != nil {
		cache.logger.Error("Failed to deserialize user status", "err", err)
		return statusFinal
	}
	return statusFinal
}
