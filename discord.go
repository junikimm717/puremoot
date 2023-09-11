package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	minValue float64 = 1.0
	zero     float64 = 0.0
	commands         = []*discordgo.ApplicationCommand{
		/*
			Broadcast commands
		*/
		{
			Name:        "broadcast",
			Description: "broadcast an anonymous message",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "The message you want to send",
					Required:    true,
				},
			},
		},
		{
			Name:        "regenerate",
			Description: "Regenerate a broadcast Id",
		},
		/*
			The namesake command
		*/
		{
			Name:        "puremoot",
			Description: "create pairings of different people",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "day",
					Description: "This is the nth puremootation",
					Required:    true,
					MinValue:    &minValue,
				},
			},
		},
		/*
			reaper commands
		*/
		{
			Name:        "reapergame",
			Description: "Command for all functions related to reaper!",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Name:        "leaderboard",
					Description: "Show the Top 20 leaderboard for a reaper game on this channel",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "gameid",
							Description: "ID of the reaper round",
							Required:    false,
							MinValue:    &minValue,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "getscore",
					Description: "Get the score of a user on a reaper round on this channel",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "User on the server",
							Required:    true,
							MinValue:    &minValue,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "gameid",
							Description: "ID of the reaper round",
							Required:    true,
							MinValue:    &minValue,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "last2reap",
					Description: "Return the last person who reaped in the current game of reaper",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "when2reap",
					Description: "Return when you can reap next",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "current",
					Description: "Show the currently active game of reaper on this channel",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "init",
					Description: "Initialize a new game of reaper (Admin Only)",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "win",
							Description: "Win Condition in seconds",
							Required:    true,
							MinValue:    &minValue,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "cooldown",
							Description: "Cooldown in seconds",
							Required:    true,
							MinValue:    &zero,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},

				{
					Name:        "cancel",
					Description: "CAREFUL!!!! This will permanently de-activate any ongoing reaper game on this channel! (Admin Only)",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "gameid",
							Description: "ID of the reaper round (confirmation purposes)",
							Required:    true,
							MinValue:    &minValue,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
		{
			Name:        "reap",
			Description: "Harvest the pears!",
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		/*
			Handlers for the reaper game
		*/
		"reap": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member == nil {
				respond(s, i, "i.Member field not present. Cannot proceed")
				return
			}
			score, err := db.Reap(i.Member.User.ID, i.ChannelID)
			if err != nil {
				respond(s, i, "Failed to Reap! "+err.Error())
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf(
						"%v reaped for %v%v Can reap again at %v",
						i.Member.User.Username,
						milliToTime(score.MilliSeconds),
						score.MultiplierMessage,
						score.ReapAgain,
					),
				},
			})
			if score.Winner != nil {
				username, err := db.UsernameFromId(*(score.Winner))
				if username == "" || err != nil {
					log.Println(err)
				}
				_, err = s.ChannelMessageSendEmbed(i.ChannelID, &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("%v Wins Reaper Round %v!", username, score.GameId),
					Description: "Thank you to everyone for playing!",
					Color:       0xFFD700,
				})
				if err != nil {
					log.Printf("Could not send embed! '%v'", err)
					return
				}
			}
		},
		"reapergame": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member == nil {
				respond(s, i, "i.Member field not present. Cannot proceed")
				return
			}
			options := i.ApplicationCommandData().Options
			handler, exists := reaperHandlers[options[0].Name]
			if !exists {
				panic("Shouldn't happen.")
			}
			handler(s, i, options[0].Options)
		},
		/*
			Commands related to broadcast functionality. People are assigned randomized ID's that they can regenerate.
		*/
		"broadcast": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := extractInteractionOptions(i.ApplicationCommandData().Options)
			pingRegex := regexp.MustCompile(`<@&?\d+>|@[0-9,a-z,A-Z]+`)
			message := options["message"].StringValue()
			if pingRegex.Match([]byte(message)) {
				respond(s, i, "Broadcast refused! Message contains a ping!")
				return
			}
			if i.Member == nil {
				respond(s, i, "[developer] The i.Member field is not present. Cannot broadcast!")
			}
			message, id := db.BroadcastMessage(i.Member.User.ID, i.ChannelID, message)
			_, err := s.ChannelMessageSend(i.ChannelID, message)
			if err != nil {
				respond(s, i, fmt.Sprintf("Could not broadcast message! '%v'", err))
			} else {
				respond(s, i, fmt.Sprintf("Message successfully broadcasted! Your ID is %v", id))
			}
		},
		"regenerate": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member == nil {
				respond(s, i, "[developer] The i.Member field is not present. Aborted")
			}
			id := db.CreateBroadcastId(i.Member.User.ID, i.ChannelID)
			respond(s, i, fmt.Sprintf("ID successfully regenerated! Your ID is %v", id))
		},
		/*
			What pureMOOt was designed to do. This command pairs people up for Randomized DM's.
		*/
		"puremoot": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !forceAdmin(s, i) {
				return
			}
			cows, err := getCows(s, i)
			if err != nil {
				respond(s, i, fmt.Sprintf("Error getting cows! %v", err))
				return
			}

			// send an embed explaining what to do
			options := extractInteractionOptions(i.ApplicationCommandData().Options)
			day := options["day"].IntValue()
			_, err = s.ChannelMessageSendEmbed(i.ChannelID, &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("pureMOOtation Day %v", day),
				Description: "pureMOOt has assigned random pairs of Cows to contact each other! Make new friends!",
				Color:       0xFFD700,
			})
			if err != nil {
				respond(s, i, fmt.Sprintf("Could not send embed! '%v'", err))
				return
			}

			respond(s, i, "pureMOOtation generated!")
			puremootation := PureMOOt(cows)
			for _, pair := range puremootation {
				num_spaces := 70 - (nickNameLength(pair[0]) + nickNameLength(pair[1]))
				prefix_spaces := 1 + rand.Intn(num_spaces-1)
				s.ChannelMessageSend(
					i.ChannelID,
					fmt.Sprintf(
						"||%v<@%v> <@%v>%v||",
						strings.Repeat(" ", 2*prefix_spaces),
						pair[0].User.ID,
						pair[1].User.ID,
						strings.Repeat(" ", 2*(num_spaces-prefix_spaces)),
					),
				)
			}
		},
	}
	reaperHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption){
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
			if !forceAdmin(s, i) {
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
			if !forceAdmin(s, i) {
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
			if !forceAdmin(s, i) {
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
)
