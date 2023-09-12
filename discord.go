package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type SubcommandHandler func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption)

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
			Type: 1,
		},
		{
			Name:        "regenerate",
			Description: "Regenerate a broadcast Id",
			Type:        1,
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
			Type: 1,
		},
		/*
			manager commands
		*/
		{

			Name:        "pmmanager",
			Description: "All commands related to the pureMOOt manager",
			Type:        1,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "set",
					Description: "Set the pureMOOt manager role",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionRole,
							Name:        "role",
							Description: "Role to set puremOOt manager to",
							Required:    true,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "unset",
					Description: "Unset the pureMOOt manager role",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "get",
					Description: "Get the name of the pureMOOt manager role",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "ballow",
					Description: "Allow broadcast on this channel (Managers Only!)",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "bban",
					Description: "Ban broadcast on this channel (Managers Only!)",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "bpermitted",
					Description: "Check if broadcast is permitted on a channel",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
		/*
			reaper commands
		*/
		{
			Name:        "reaper",
			Description: "Command for all functions related to reaper!",
			Type:        1,
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
					Name:        "score",
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
			Type:        1,
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
						"%v reaped for %v%v They will not be able to reap until %v",
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
		"reaper": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member == nil {
				respond(s, i, "i.Member field not present. Cannot proceed")
				return
			}
			options := i.ApplicationCommandData().Options
			handler, exists := ReaperHandlers[options[0].Name]
			if !exists {
				respond(s, i, "Invalid Command given")
				return
			}
			handler(s, i, options[0].Options)
		},
		/*
			Commands related to managing roles, etc.
		*/
		"pmmanager": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member == nil {
				respond(s, i, "i.Member field not present. Cannot proceed")
				return
			}
			options := i.ApplicationCommandData().Options
			handler, exists := ManagerHandlers[options[0].Name]
			if !exists {
				respond(s, i, "Invalid Command given")
				return
			}
			handler(s, i, options[0].Options)
		},
		/*
			Commands related to broadcast functionality. People are assigned randomized ID's that they can regenerate.
		*/
		"broadcast": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := extractInteractionOptions(i.ApplicationCommandData().Options)
			if !db.BroadcastAllowed(i.ChannelID) {
				respond(s, i, "Anonymous broadcasting has been banned on this channel!")
				return
			}
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
			if !forceManager(s, i) {
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
)
