package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Database struct {
	client *redis.Client
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
	}
}

func (d *Database) GetString(key string) (string, bool) {
	val, err := d.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", false
	} else if err != nil {
		panic(err)
	}
	return val, true
}

func (d *Database) SetString(key string, val string) {
	err := d.client.Set(ctx, key, val, 0).Err()
	if err != nil {
		panic(err)
	}
}

func (d *Database) UsernameFromId(id string) (string, error) {
	username, exists := d.GetString(fmt.Sprintf("userid:%v", id))
	if !exists {
		user, err := dg.User(id)
		if err != nil {
			return "<nonexistent user>", err
		}
		username = user.Username
		cache_time, err := time.ParseDuration("24h")
		// should never happen.
		if err != nil {
			panic(err)
		}
		d.client.Set(ctx, fmt.Sprintf("userid:%v", id), username, cache_time)
	}
	return username, nil
}
