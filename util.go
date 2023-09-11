package main

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func nickNameLength(m *discordgo.Member) int {
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

func extractInteractionOptions(options []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	res := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		res[opt.Name] = opt
	}
	return res
}

func milliToTime(milliseconds int64) string {
	res := []string{}
	t := milliseconds / 1000
	type timeSegment struct {
		Length int64
		Name   string
	}
	if t == 0 {
		return "0 seconds"
	}
	timesegments := []timeSegment{
		{Length: 24 * 3600 * 7, Name: "week"},
		{Length: 24 * 3600, Name: "day"},
		{Length: 3600, Name: "hour"},
		{Length: 60, Name: "minute"},
		{Length: 1, Name: "second"},
	}
	for _, seg := range timesegments {
		x := t / seg.Length
		t = t % seg.Length
		if x == 1 {
			res = append(res, strconv.FormatInt(x, 10)+" "+seg.Name)
		} else if x > 1 {
			res = append(res, strconv.FormatInt(x, 10)+" "+seg.Name+"s")
		}
	}
	return strings.Join(res, ", ")
}
