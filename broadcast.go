package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

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

func (d *Database) userChannelKey(user string, channel string) string {
	return fmt.Sprintf("%v-%v", user, channel)
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

/*
Functions for enabling and disabling broadcast
*/
func (d *Database) EnableBroadcast(channel string) error {
	return d.SetBool(
		fmt.Sprintf("broadcast-enabled:%v", channel),
		true,
	)
}

func (d *Database) DisableBroadcast(channel string) error {
	return d.SetBool(
		fmt.Sprintf("broadcast-enabled:%v", channel),
		false,
	)
}

func (d *Database) BroadcastAllowed(channel string) bool {
	allowed, exists := d.GetBool(
		fmt.Sprintf("broadcast-enabled:%v", channel),
	)
	if !exists {
		return false
	}
	return allowed
}
