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
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "respondto",
					Description: "Respond to the message with this link in a new thread",
					Required:    false,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "threadname",
					Description: "Name of the new thread",
					MaxLength:   15,
					Required:    false,
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
					Name:        "unset",
					Description: "Unset the pureMOOt manager role",
					Options: []*discordgo.ApplicationCommandOption{
						&PuremootRoleOptions,
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "get",
					Description: "Get the name of the pureMOOt manager role",
					Options: []*discordgo.ApplicationCommandOption{
						&PuremootRoleOptions,
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "set",
					Description: "Set the pureMOOt manager role",
					Options: []*discordgo.ApplicationCommandOption{
						&PuremootRoleOptions,
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
					Name: "rban", Description: "Ban a user from reaping on this channel (Admin Only)",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to ban",
							Required:    true,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name: "rallow", Description: "Allow a user to reap on this channel (Admin Only)",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to allow",
							Required:    true,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name: "rpermitted", Description: "Check if a user is permitted to reap on this channel",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionUser,
							Name:        "user",
							Description: "The user to check",
							Required:    true,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
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
					Name:        "last",
					Description: "Return the last person who reaped in the current game of reaper",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "when",
					Description: "Shows when you can reap next",
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "active",
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
			Description: "Harvest a pear!",
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
			if db.IsReaperUserBanned(i.Member.User.ID, i.ChannelID) {
				respond(s, i, "The Admins have banned you from reaping!")
				return
			}
			score, err := db.Reap(i.Member.User.ID, i.ChannelID)
			if err != nil {
				respond(s, i, "Failed to Reap! "+err.Error())
				return
			}

			freeReapMessageComps := []string{}
			if score.FreeReap {
				freeReapMessageComps = append(freeReapMessageComps, "Free Reap Gained!")
			}
			if score.FreeReapUsed {
				freeReapMessageComps = append(freeReapMessageComps, "Used a Free Reap!")
			}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf(
						"%v reaped for %v%v Their cooldown expires at %v",
						i.Member.User.Username,
						milliToTime(score.MilliSeconds),
						score.MultiplierMessage,
						score.ReapAgain,
					) + "\n" + strings.Join(freeReapMessageComps, " "),
				},
			})
			if score.Winner != nil {
				username, err := db.UsernameFromId(*(score.Winner))
				if username == "" || err != nil {
					log.Println(err)
				}
				leaderboard, err := db.GetLeaderBoard(i.ChannelID, score.GameId)
				if err != nil {
					s.ChannelMessageSend(
						i.ChannelID,
						fmt.Sprintf("Could not send embed! '%v'", err),
					)
					return
				}
				medals := []string{":first_place:", ":second_place:", ":third_place:"}
				usernames := []string{}
				ranks := []string{}
				scores := []string{}
				for rank, item := range leaderboard {
					if rank < len(medals) {
						ranks = append(ranks, medals[rank])
					} else {
						ranks = append(ranks, fmt.Sprintf("%v", rank+1))
					}
					usernames = append(usernames, fmt.Sprintf("%v", item.Username))
					scores = append(scores, fmt.Sprintf("%.3f seconds", item.Score))
				}

				channelname, _ := db.ChannelFromId(i.ChannelID)
				_, err = s.ChannelMessageSendEmbed(i.ChannelID, &discordgo.MessageEmbed{
					Title: fmt.Sprintf(
						"%v wins Reaper Round %v in #%v!",
						username,
						score.GameId,
						channelname,
					),
					Description: "Thank you everyone for playing!",

					Fields: []*discordgo.MessageEmbedField{
						{Name: "Rank", Value: strings.Join(ranks, "\n"), Inline: true},
						{Name: "Username", Value: strings.Join(usernames, "\n"), Inline: true},
						{Name: "Score", Value: strings.Join(scores, "\n"), Inline: true},
					},
					Color: 0xFFD700,
				})

				if err != nil {
					s.ChannelMessageSend(
						i.ChannelID,
						fmt.Sprintf("Could not send embed! '%v'", err),
					)
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
			if i.Member == nil {
				respond(s, i, "[developer] The i.Member field is not present. Cannot broadcast!")
			}
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

			respondoption, exists := options["respondto"]
			if !exists {
				broadcastMessage, id := db.BroadcastMessage(i.Member.User.ID, i.ChannelID, message)
				_, err := s.ChannelMessageSend(i.ChannelID, broadcastMessage)
				if err != nil {
					respond(s, i, fmt.Sprintf("Could not broadcast message! '%v'", err))
				} else {
					respond(s, i, fmt.Sprintf("Message successfully broadcasted! Your ID is %v", id))
				}
				return
			}
			respondto := respondoption.StringValue()
			messageLinkRegex := regexp.MustCompile(`https?://discord.com/channels/(?P<guildId>\d+)/(?P<channelId>\d+)/(?P<messageId>\d+)`)
			match := messageLinkRegex.FindStringSubmatch(respondto)

			if len(match) != 4 {
				respond(s, i, fmt.Sprintf("Invalid Message Link!"))
				return
			}

			if i.ChannelID != match[2] {
				matchchannel, err := db.ChannelFromId(match[2])
				if err != nil {
					respond(s, i, err.Error())
					return
				}
				currentchannel, err := db.ChannelFromId(i.ChannelID)
				if err != nil {
					respond(s, i, err.Error())
					return
				}
				respond(s, i, fmt.Sprintf("Message belongs to channel %v but you are in %v", matchchannel, currentchannel))
				return
			}
			messagetorespond, err := dg.ChannelMessage(match[2], match[3])
			if err != nil {
				respond(s, i, fmt.Sprintf(err.Error()))
				return
			} else if messagetorespond == nil {
				respond(s, i, "The message you were trying to respond to is nil!")
			}

			threadnameoption, exists := options["threadname"]
			threadname := ""
			if exists {
				threadname = threadnameoption.StringValue()
			} else {
				threadname = truncate(messagetorespond.ContentWithMentionsReplaced(), 15)
			}

			thread, err := s.MessageThreadStart(
				messagetorespond.ChannelID,
				messagetorespond.ID,
				fmt.Sprintf(
					"pureMOOt-%v",
					threadname,
				),
				60*24*3,
			)
			if err != nil {
				respond(s, i, fmt.Sprintf("Failed to create new thread! %v", err.Error()))
				return
			}

			broadcastMessage, id := db.BroadcastMessage(i.Member.User.ID, thread.ID, message)
			_, err = s.ChannelMessageSend(thread.ID, broadcastMessage)
			if err != nil {
				respond(s, i, fmt.Sprintf("Could not broadcast message! '%v'", err))
			} else {
				respond(s, i,
					fmt.Sprintf(
						"Message successfully broadcasted into thread 'pureMOOt-%v'! Your ID is %v",
						threadname,
						id,
					),
				)
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
