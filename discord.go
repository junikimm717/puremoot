package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Pair[T any] struct {
	Key   float64
	Value T
}

// TODO: implement this
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

func NickNameLength(m *discordgo.Member) int {
	if len(m.Nick) > 0 {
		return len(m.Nick)
	} else {
		return len(m.User.Username)
	}
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func extractInteractionOptions(i *discordgo.InteractionCreate) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	options := i.ApplicationCommandData().Options
	res := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		res[opt.Name] = opt
	}
	return res
}

func getCows(s *discordgo.Session, i *discordgo.InteractionCreate) ([]*discordgo.Member, error) {
	members, err := s.GuildMembers(i.GuildID, "", 300)
	if err != nil {
		return []*discordgo.Member{}, err
	}
	cow_role_id := ""
	roles, err := s.GuildRoles(i.GuildID)
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

func ForceAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Member == nil {
		respond(s, i, "You cannot invoke this command outside of a guild")
		return false
	}
	if i.Member.Permissions&(discordgo.PermissionAdministrator|discordgo.PermissionManageServer) != 0 {
		return true
	}
	respond(s, i, "You do not have an administrator role! Permission Denied")
	return false
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
			Name:        "cows",
			Description: "commmand to see all cows on the server",
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"broadcast": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := extractInteractionOptions(i)
			_, err := s.ChannelMessageSend(i.ChannelID, "[broadcast] "+options["message"].StringValue())
			if err != nil {
				respond(s, i, fmt.Sprintf("Could not broadcast message! '%v'", err))
			} else {
				respond(s, i, "Message successfully broadcasted!")
			}
		},
		"cows": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !ForceAdmin(s, i) {
				return
			}
			cows, err := getCows(s, i)
			if err != nil {
				respond(s, i, fmt.Sprintf("Error getting cows! %v", err))
			}
			cowids := make([]string, len(cows))
			for idx, cow := range cows {
				cowids[idx] = cow.User.ID
			}
			respond(s, i, fmt.Sprintf("Cows: %v", strings.Join(cowids, ",")))
		},
		"puremoot": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if !ForceAdmin(s, i) {
				return
			}
			cows, err := getCows(s, i)
			if err != nil {
				respond(s, i, fmt.Sprintf("Error getting cows! %v", err))
			}
			options := extractInteractionOptions(i)
			day := options["day"].IntValue()
			s.ChannelMessageSendEmbed(i.ChannelID, &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("PureMOOtation day %v", day),
				Description: "PureMOOt has assigned random pairs of MOOpers to contact each other! Make New Friends!",
				Color:       0xFFD700,
			})
			puremootation := PureMOOt(cows)
			for _, pair := range puremootation {
				num_spaces := 70 - (NickNameLength(pair[0]) + NickNameLength(pair[1]))
				prefix_spaces := 1 + rand.Intn(num_spaces-1)
				s.ChannelMessageSend(
					i.ChannelID,
					fmt.Sprintf(
						"||%v<@%v> <@%v>%v||",
						strings.Repeat(" ", prefix_spaces),
						pair[0].User.ID,
						pair[1].User.ID,
						strings.Repeat(" ", num_spaces-prefix_spaces),
					),
				)
			}
			respond(s, i, "Success.")
		},
	}
)
