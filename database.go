package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var letters = []byte{}

func init() {
	for i := 0; i < 26; i++ {
		letters = append(letters, byte(65+i))
		letters = append(letters, byte(97+i))
	}
}

func randomId() string {
	id := []byte{}
	for i := 0; i < 6; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			panic(err)
		}
		id = append(id, letters[nBig.Int64()])
	}
	return string(id)
}

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

func (d *Database) userChannelKey(user string, channel string) string {
	return fmt.Sprintf("%v-%v", user, channel)
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

func (d *Database) CreateBroadcastId(user string, channel string) string {
	id := randomId()
	d.SetString(d.userChannelKey(user, channel), id)
	return id
}

func (d *Database) BroadcastMessage(user string, channel string, message string) (string, string) {
	id, exists := d.GetString(d.userChannelKey(user, channel))
	if !exists {
		id = d.CreateBroadcastId(user, channel)
	}
	return fmt.Sprintf("[broadcast:**%v**] %v", id, message), id
}
