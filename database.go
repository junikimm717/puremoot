package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type Database struct {
	client *redis.Client
	ctx    context.Context
}

func InitDatabase() *Database {
	addr, set := os.LookupEnv("PUREMOOT_REDIS")
	if !set {
		addr = "localhost:6379"
		log.Println("PUREMOOT_REDIS environment variable not found, using localhost:6379...")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &Database{
		client: rdb,
		ctx:    context.Background(),
	}
}

func (d *Database) GetString(key string) (string, bool) {
	val, err := d.client.Get(d.ctx, key).Result()
	if err == redis.Nil {
		return "", false
	} else if err != nil {
		panic(err)
	}
	return val, true
}

func (d *Database) SetString(key string, val string) error {
	return d.client.Set(d.ctx, key, val, 0).Err()
}

func (d *Database) GetBool(key string) (bool, bool) {
	val, err := d.client.Get(d.ctx, key).Bool()
	if err == redis.Nil {
		return false, false
	} else if err != nil {
		panic(err)
	}
	return val, true
}

func (d *Database) SetBool(key string, val bool) error {
	return d.client.Set(d.ctx, key, val, 0).Err()
}

func (d *Database) GetInt(key string) (int, bool) {
	val, err := d.client.Get(d.ctx, key).Int()
	if err == redis.Nil {
		return 0, false
	} else if err != nil {
		panic(err)
	}
	return val, true
}

func (d *Database) SetInt(key string, val int) error {
	return d.client.Set(d.ctx, key, val, 0).Err()
}

func (d *Database) UsernameFromId(userId string) (string, error) {
	username, exists := d.GetString(fmt.Sprintf("userid:%v", userId))
	number, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		// should never happen.
		panic(err)
	}
	// randomly re-validate usernames 8% of the time
	if !exists || number.Int64() < int64(80) {
		user, err := dg.User(userId)
		if err != nil {
			return "<nonexistent user>", err
		}
		username = user.Username
		cache_time, err := time.ParseDuration("24h")
		// should never happen.
		if err != nil {
			panic(err)
		}
		d.client.Set(d.ctx, fmt.Sprintf("userid:%v", userId), username, cache_time)
	}
	return username, nil
}

func (d *Database) ChannelFromId(channelId string) (string, error) {
	channelname, exists := d.GetString(fmt.Sprintf("channelid:%v", channelId))
	number, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		// should never happen.
		panic(err)
	}
	// randomly re-validate usernames 8% of the time
	if !exists || number.Int64() < int64(150) {
		channel, err := dg.Channel(channelId)
		if err != nil {
			return "<nonexistent channel>", err
		}
		channelname = channel.Name
		cache_time, err := time.ParseDuration("24h")
		// should never happen.
		if err != nil {
			panic(err)
		}
		d.client.Set(d.ctx, fmt.Sprintf("channelid:%v", channelId), channelname, cache_time)
	}
	return channelname, nil
}
