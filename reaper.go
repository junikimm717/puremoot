package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
)

/*
Database schema goes as follows:
reaper:{channel}:latest - returns the id of the most recent game.
reaper:{channel}:active - returns whether the most recent game is running.

reaper:{channel}:{id}:wincond - number of seconds needed to win (int64)
reaper:{channel}:{id}:winner - a string that contains winner (nil if not set)
reaper:{channel}:{id}:cooldown - minimum amount of time between reaps (float64)
reaper:{channel}:{id}:leaderboard - an ordered set that contains reaper scores.
reaper:{channel}:{id}:reaplog - a stream that contains the most recent reaps.
reaper:{channel}:{id}:{user}:last - when did this user last reap?
*/

type ReapLogEntry struct {
	UserId string
}

/*
return the latest reaper game on the channel and also whether it is currently
running or not.
*/
func (d *Database) CurrentReaperId(channelid string) (int, bool) {
	val, err := d.client.Get(ctx, fmt.Sprintf("reaper:%v:latest", channelid)).Int()
	if err == redis.Nil {
		return 0, false
	} else if err != nil {
		panic(err)
	}
	running, err := d.client.Get(ctx, fmt.Sprintf("reaper:%v:active", channelid)).Bool()
	if err == redis.Nil {
		return val, false
	} else if err != nil {
		panic(err)
	}
	return val, running
}

func (d *Database) LastReapTimeForUser(channelid string, gameid int, userid string) (int64, bool) {
	last_reaped, err := d.client.Get(
		ctx,
		fmt.Sprintf("reaper:%v:%v:%v:last", channelid, gameid, userid),
	).Int64()
	if err == redis.Nil {
		return 0, false
	} else if err != nil {
		log.Fatalln(err)
	}
	return last_reaped, true
}

func (d *Database) LastToReap(channelid string, gameid int) (string, int64) {
	last, err := d.client.XRevRangeN(
		ctx,
		fmt.Sprintf("reaper:%v:%v:reaplog", channelid, gameid),
		"+",
		"-",
		1,
	).Result()
	if err != nil {
		log.Fatalln(err)
	}
	lastreap, err := strconv.ParseInt(last[0].ID[:len(last[0].ID)-2], 10, 64)
	if err != nil {
		log.Fatalln(err)
	}
	return last[0].Values["userid"].(string), lastreap
}

func (d *Database) GetCooldown(channelid string, gameid int) int64 {
	val, err := d.client.Get(ctx, fmt.Sprintf("reaper:%v:%v:cooldown", channelid, gameid)).Int64()
	if err != nil {
		panic(err)
	}
	return val
}

func (d *Database) GetWincond(channelid string, gameid int) int64 {
	wincond, err := d.client.Get(ctx, fmt.Sprintf("reaper:%v:%v:wincond", channelid, gameid)).Int64()
	if err != nil {
		panic(err)
	}
	return wincond
}

func (d *Database) GetTheLeader(channelid string, gameid int) *redis.Z {
	// take the first from the leaderboard
	leader, err := d.client.ZRevRangeWithScores(
		ctx,
		fmt.Sprintf("reaper:%v:%v:leaderboard", channelid, gameid),
		0,
		0,
	).Result()
	if err == redis.Nil || len(leader) == 0 {
		return nil
	} else if err != nil {
		panic(err)
	}
	return &leader[0]
}

type LeaderBoardItem struct {
	Username string
	Score    float64
}

/*
Shows the leaderboard for the requested game of reaper.
Returns the leaderboard, the game item used, and any errors.
*/
func (d *Database) GetLeaderBoard(channelid string, gameid int) ([]LeaderBoardItem, error) {
	// take the first from the leaderboard
	leaders, err := d.client.ZRevRangeWithScores(
		ctx,
		fmt.Sprintf("reaper:%v:%v:leaderboard", channelid, gameid),
		0,
		19,
	).Result()
	if err == redis.Nil || len(leaders) == 0 {
		return []LeaderBoardItem{}, nil
	} else if err != nil {
		panic(err)
	}
	res := make([]LeaderBoardItem, 0)
	for _, leader := range leaders {
		username, err := d.UsernameFromId(leader.Member.(string))
		if username != "" && err == nil {
			res = append(res, LeaderBoardItem{
				Username: username,
				Score:    leader.Score,
			})
		}
	}
	return res, nil
}

type RankResponse struct {
	Score float64
	Rank  int64
}

func (d *Database) GetOneScore(channelid string, gameid int, userid string) (RankResponse, error) {
	currentid, _ := db.CurrentReaperId(channelid)
	if gameid > currentid || gameid == 0 {
		return RankResponse{}, errors.New("No such game found!")
	}
	rank, err := d.client.ZRevRank(
		ctx,
		fmt.Sprintf("reaper:%v:%v:leaderboard", channelid, gameid),
		userid,
	).Result()

	if err == redis.Nil {
		return RankResponse{}, errors.New("Hmm...it seems this user did not participate in this round of reaper.")
	} else if err != nil {
		panic(err)
	}

	scores, err := d.client.ZMScore(
		ctx,
		fmt.Sprintf("reaper:%v:%v:leaderboard", channelid, gameid),
		userid,
	).Result()

	if err == redis.Nil || len(scores) == 0 {
		return RankResponse{}, errors.New("Hmm...it seems this user did not participate in this round of reaper.")
	} else if err != nil {
		panic(err)
	}

	return RankResponse{Rank: rank + 1, Score: scores[0]}, nil
}

/*
Initialize a game of reaper on a channel
Parameters:
channel - id of the discord channel where the game will be played.
win - number of seconds to win (int)
cooldown - number of seconds between reaps.

1. Make sure that the discord channel doesn't have any other games of reaper going on.
2. Create a leaderboard (key is user id and value is their score)
3. Store reaper log in the form of a Redis stream.
4. Track when each player last reaped for cooldown.

Once this is done, the timer should start.
*/
func (d *Database) InitReaper(channelid string, wincond int64, cooldown int64) (int, bool) {
	newgameid, running := d.CurrentReaperId(channelid)
	if running {
		return 0, false
	}
	newgameid += 1
	err := d.client.Set(ctx, fmt.Sprintf("reaper:%v:active", channelid), true, 0).Err()
	if err != nil {
		log.Fatalln(err)
	}
	err = d.client.Set(ctx, fmt.Sprintf("reaper:%v:latest", channelid), newgameid, 0).Err()
	if err != nil {
		log.Fatalln(err)
	}
	err = d.client.Set(ctx, fmt.Sprintf("reaper:%v:%v:wincond", channelid, newgameid), wincond, 0).Err()
	if err != nil {
		log.Fatalln(err)
	}
	err = d.client.Set(ctx, fmt.Sprintf("reaper:%v:%v:cooldown", channelid, newgameid), cooldown, 0).Err()
	if err != nil {
		log.Fatalln(err)
	}
	err = d.client.XAdd(ctx, &redis.XAddArgs{
		Stream: fmt.Sprintf("reaper:%v:%v:reaplog", channelid, newgameid),
		Values: map[string]interface{}{
			"userid": "puremoot",
		},
		MaxLen: 200,
	}).Err()
	if err != nil {
		log.Fatalln(err)
	}
	return newgameid, true
}

/*
Cancel the current game of reaper.
*/
func (d *Database) CancelReaper(channelid string) (int, bool) {
	gameid, running := d.CurrentReaperId(channelid)
	if !running {
		return 0, false
	}
	d.client.Set(ctx, fmt.Sprintf("reaper:%v:active", channelid), false, 0)
	return gameid, true
}

func (d *Database) When2Reap(userid string, channelid string) (int64, error) {
	gameid, running := d.CurrentReaperId(channelid)
	if !running {
		return 0, errors.New("There is no active game of reaper on this channel. Ask the admins.")
	}
	userlastreap, didreap := d.LastReapTimeForUser(channelid, gameid, userid)
	cooldown := d.GetCooldown(channelid, gameid)
	if didreap {
		return (userlastreap)/1000 + cooldown, nil
	} else {
		return 0, nil
	}
}

func Multiplier() (int64, string) {
	number, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		panic(err)
	}
	multipliers := []struct {
		Prob    int
		Mult    int64
		Message string
	}{
		// all probabilities are out of 1000
		{
			Prob:    1,
			Mult:    8,
			Message: "**Ultra Rare Octuple Reap!**",
		},
		{
			Prob:    5,
			Mult:    5,
			Message: "**Rare Quintuple Reap!**",
		},
		{
			Prob:    10,
			Mult:    4,
			Message: "**Quadruple Reap!**",
		},
		{
			Prob:    40,
			Mult:    3,
			Message: "**Triple Reap!**",
		},
		{
			Prob:    40,
			Mult:    2,
			Message: "**Double Reap!**",
		},
	}
	sum := 0
	for _, m := range multipliers {
		if sum <= int(number.Int64()) && int(number.Int64()) < sum+m.Prob {
			return m.Mult, " " + m.Message
		}
		sum += m.Prob
	}
	return 1, "."
}

type ReapOutput struct {
	MilliSeconds      int64
	ReapAgain         string
	MultiplierMessage string
	Winner            *string
	GameId            int
}

func (d *Database) Reap(userid string, channelid string) (ReapOutput, error) {
	gameid, running := d.CurrentReaperId(channelid)
	if !running {
		return ReapOutput{}, errors.New("There is no active game of reaper. Ask the admins.")
	}
	// removed for testing purposes.
	//_, lastreaptime := d.LastToReap(channelid, gameid)
	lastreaper, lastreaptime := d.LastToReap(channelid, gameid)
	if userid == lastreaper {
		return ReapOutput{}, errors.New("You were the last person to reap.")
	}

	/*
		!!!!! userlastreap is in unix milliseconds
	*/
	userlastreap, didreap := d.LastReapTimeForUser(channelid, gameid, userid)
	cooldown := d.GetCooldown(channelid, gameid)
	if didreap {
		if userlastreap+cooldown*1000 > time.Now().UnixMilli() {
			return ReapOutput{}, fmt.Errorf(
				"You may not reap again until %v",
				fmt.Sprintf("<t:%v>", (userlastreap)/1000+cooldown),
			)
		}
	}
	// log that this was the last time that this user reaped.
	err := d.client.Set(
		ctx,
		fmt.Sprintf("reaper:%v:%v:%v:last", channelid, gameid, userid),
		time.Now().UnixMilli(),
		0,
	).Err()
	if err != nil {
		log.Fatalln(err)
	}

	// calculate reaper score
	timenow := time.Now()
	score := timenow.UnixMilli() - lastreaptime

	message := "." // a period is included because some messages have punctuation.
	if score < 3600*1000 {
		multiplier, m := Multiplier()
		score *= multiplier
		message = m
	}

	// add it to the streams.
	err = d.client.XAdd(ctx, &redis.XAddArgs{
		Stream: fmt.Sprintf("reaper:%v:%v:reaplog", channelid, gameid),
		Values: map[string]interface{}{
			"userid": userid,
		},
	}).Err()
	if err != nil {
		log.Fatalln(err)
	}
	// add it to the scoreboard.
	err = d.client.ZAddArgsIncr(
		ctx,
		fmt.Sprintf("reaper:%v:%v:leaderboard", channelid, gameid),
		redis.ZAddArgs{
			Members: []redis.Z{
				{
					Score:  (float64)(score) / 1000,
					Member: userid,
				},
			},
		},
	).Err()
	if err != nil {
		log.Fatalln(err)
	}
	leader := d.GetTheLeader(channelid, gameid)
	wincond := d.GetWincond(channelid, gameid)
	if leader != nil {
		if leader.Score >= float64(wincond) {
			err := d.client.Set(ctx, fmt.Sprintf("reaper:%v:active", channelid), false, 0).Err()
			if err != nil {
				log.Fatalln(err)
			}
			winner := leader.Member.(string)
			return ReapOutput{
				MilliSeconds:      score,
				Winner:            &winner,
				ReapAgain:         fmt.Sprintf("<t:%v>", timenow.Unix()+cooldown),
				GameId:            gameid,
				MultiplierMessage: message,
			}, nil
		}
	}
	return ReapOutput{
		MilliSeconds:      score,
		ReapAgain:         fmt.Sprintf("<t:%v>", timenow.Unix()+cooldown),
		GameId:            gameid,
		MultiplierMessage: message,
	}, nil
}

var ReaperHandlers = map[string]SubcommandHandler{
	"leaderboard": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		options := extractInteractionOptions(opts)
		currentid, activegame := db.CurrentReaperId(i.ChannelID)
		var leaderboard []LeaderBoardItem
		leaderboardgameid := int64(0)
		if gameid, ok := options["gameid"]; ok {
			if currentid == 0 || gameid.IntValue() > int64(currentid) {
				respond(s, i, "No such round of Reaper!")
				return
			}
			respond(s, i, "Message Received! Compiling Leaderboard...")
			b, err := db.GetLeaderBoard(i.ChannelID, int(gameid.IntValue()))
			if err != nil {
				s.ChannelMessageSend(
					i.ChannelID,
					"Error generating Leaderboard! "+err.Error(),
				)
				return
			}
			leaderboard = b
			leaderboardgameid = gameid.IntValue()
		} else {
			if !activegame {
				respond(s, i, "No active game of reaper!")
				return
			}
			respond(s, i, "Message Received! Compiling Leaderboard...")
			b, err := db.GetLeaderBoard(i.ChannelID, currentid)
			if err != nil {
				s.ChannelMessageSend(
					i.ChannelID,
					"Error generating Leaderboard! "+err.Error(),
				)
				return
			}
			leaderboard = b
			leaderboardgameid = int64(currentid)
		}

		wincond := db.GetWincond(i.ChannelID, int(leaderboardgameid))
		cooldown := db.GetCooldown(i.ChannelID, int(leaderboardgameid))

		usernames := []string{}
		ranks := []string{}
		scores := []string{}
		for rank, item := range leaderboard {
			ranks = append(ranks, fmt.Sprintf("%v.", rank+1))
			usernames = append(usernames, fmt.Sprintf("%v", item.Username))
			scores = append(scores, fmt.Sprintf("%.3f seconds", item.Score))
		}

		_, err := s.ChannelMessageSendEmbed(i.ChannelID, &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Reaper Round %v", leaderboardgameid),
			Description: fmt.Sprintf("The Top 20 Leaderboard | **%v** seconds to win | **%v** seconds between reaps", wincond, cooldown),
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Rank", Value: strings.Join(ranks, "\n"), Inline: true},
				{Name: "Username", Value: strings.Join(usernames, "\n"), Inline: true},
				{Name: "Score", Value: strings.Join(scores, "\n"), Inline: true},
			},
			Color: 0xFFD700,
		})
		if err != nil {
			respond(s, i, fmt.Sprintf("Error Sending Message! %v", err.Error()))
		}
	},
	"score": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		options := extractInteractionOptions(opts)
		user := options["user"].UserValue(s)
		gameid := options["gameid"].IntValue()
		score, err := db.GetOneScore(i.ChannelID, int(gameid), user.ID)
		if err != nil {
			respond(s, i, err.Error())
		}
		respond(
			s,
			i,
			fmt.Sprintf("%v reaped a total of %v seconds (Rank %v) in Round %v!", user.Username, score.Score, score.Rank, gameid),
		)
	},
	"init": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		options := extractInteractionOptions(opts)
		win := options["win"].IntValue()
		cooldown := options["cooldown"].IntValue()
		gameId, created := db.InitReaper(i.ChannelID, win, cooldown)
		if !created {
			respond(s, i, "The ongoing reaper round must end before you can create a new round!")
			return
		}
		_, err := s.ChannelMessageSendEmbed(i.ChannelID, &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("Reaper Round %v Has Begun!", gameId),
			Description: fmt.Sprintf("%v seconds to win. %v seconds between reaps. Use the `/reap` command.", win, cooldown),
			Color:       0xFFD700,
		})
		if err != nil {
			respond(s, i, fmt.Sprintf("Error Sending Message! %v", err))
		} else {
			respond(s, i, fmt.Sprintf("Reaper Round %v Successfully Created", gameId))
		}
	},
	"cancel": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		options := extractInteractionOptions(opts)
		gameid, _ := db.CurrentReaperId(i.ChannelID)
		if gameid != int(options["gameid"].IntValue()) {
			respond(s, i, "Confirmation failed. Game has not been cancelled.")
			return
		}
		gameid, deleted := db.CancelReaper(i.ChannelID)
		if !deleted {
			respond(s, i, "There is no active game of reaper!")
			return
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Reaper Round %v has been cancelled by the admins! Very 1428!", gameid),
			},
		})
	},
	"current": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		gameId, active := db.CurrentReaperId(i.ChannelID)
		if !active {
			if gameId == 1 {
				respond(s, i, fmt.Sprintf("There is no active game of reaper on this channel! %v round has been played.", gameId))
			} else {
				respond(s, i, fmt.Sprintf("There is no active game of reaper on this channel! %v rounds have been played.", gameId))
			}
			return
		}
		respond(s, i, fmt.Sprintf("Reaper Round %v is active!", gameId))
	},
	"last2reap": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		gameid, active := db.CurrentReaperId(i.ChannelID)
		if !active {
			if gameid == 1 {
				respond(s, i, fmt.Sprintf("There is no active game of reaper on this channel! %v round has been played.", gameid))
			} else {
				respond(s, i, fmt.Sprintf("There is no active game of reaper on this channel! %v rounds have been played.", gameid))
			}
			return
		}
		lastreaper, lastreaptime := db.LastToReap(i.ChannelID, gameid)
		username, err := db.UsernameFromId(lastreaper)
		if err != nil {
			log.Println(err)
		}
		respond(
			s,
			i,
			fmt.Sprintf("%v last reaped at <t:%v>", username, lastreaptime/1000),
		)
	},
	"when2reap": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		time, err := db.When2Reap(i.Member.User.ID, i.ChannelID)
		if err != nil {
			respond(s, i, err.Error())
		}
		if time == 0 {
			respond(s, i, "You haven't reaped yet. Go ahead!")
		} else {
			respond(
				s,
				i,
				fmt.Sprintf("You may reap at <t:%v>", time),
			)
		}
	},
}
