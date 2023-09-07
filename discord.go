package main

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Pair[T any] struct {
	Key   float64
	Value T
}

func PureMOOt[T any](arr []T) [][]T {
	pairs := make([]Pair[T], len(arr))
	for idx := 0; idx < len(pairs); idx++ {
		pairs[idx] = Pair[T]{Key: rand.Float64(), Value: arr[idx]}
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Key < pairs[j].Key
	})
	res := make([][]T, 0)
	for idx := 0; idx+1 < len(pairs); idx += 2 {
		res = append(res, []T{pairs[idx].Value, pairs[idx+1].Value})
	}
	return res
}

func getCows(s *discordgo.Session, i *discordgo.InteractionCreate) ([]*discordgo.Member, error) {
	cow_role_id := ""
	roles, err := s.GuildRoles(i.GuildID)
	if err != nil {
		return []*discordgo.Member{}, err
	}
	members, err := s.GuildMembers(i.GuildID, "", 800)
	if err != nil {
		return []*discordgo.Member{}, err
	}
	for _, role := range roles {
		if strings.ToLower(role.Name) == "cow" {
			cow_role_id = role.ID
			break
		}
	}
	if cow_role_id == "" {
		return []*discordgo.Member{}, errors.New("No Server Role 'cow'.")
	}
	cows := make([]*discordgo.Member, 0)
	for _, m := range members {
		for _, role := range m.Roles {
			if role == cow_role_id {
				cows = append(cows, m)
			}
		}
	}
	return cows, nil
}

var (
	minValue float64 = 1.0
	commands         = []*discordgo.ApplicationCommand{
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
			Name:        "mybroadcastid",
			Description: "Returns your broadcast ID on this channel",
		},
		{
			Name:        "regenerate",
			Description: "Regenerate a broadcast Id",
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"broadcast": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := extractInteractionOptions(i)
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
		"mybroadcastid": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if i.Member == nil {
				respond(s, i, "[developer] The i.Member field is not present. Aborted")
			}
			id, exists := db.GetString(db.userChannelKey(i.Member.User.ID, i.ChannelID))
			if !exists {
				respond(s, i, fmt.Sprintf("You have not been assigned a Broadcast ID on this channel."))
			} else {
				respond(s, i, fmt.Sprintf("Your ID is %v", id))
			}
		},
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
			options := extractInteractionOptions(i)
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
