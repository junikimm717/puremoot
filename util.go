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

func forceAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Member == nil {
		respond(s, i, "You cannot invoke this command outside of a guild")
		return false
	}
	manage_audit := int64(discordgo.PermissionManageServer | discordgo.PermissionViewAuditLogs)
	if (i.Member.Permissions&(discordgo.PermissionAdministrator) != 0) || (i.Member.Permissions&manage_audit) == manage_audit {
		return true
	}
	respond(s, i, "You do not have an administrator role! Permission Denied")
	return false
}

func milliToTime(milliseconds int64) string {
	res := []string{}
	t := milliseconds / 1000
	type timesegment struct {
		Length int64
		Name   string
	}
	if t == 0 {
		return "0 seconds"
	}
	timesegments := []timesegment{
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
