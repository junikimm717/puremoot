package main

import (
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

func extractInteractionOptions(i *discordgo.InteractionCreate) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	options := i.ApplicationCommandData().Options
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
  manage_audit := int64(discordgo.PermissionManageServer|discordgo.PermissionViewAuditLogs)
	if (i.Member.Permissions&(discordgo.PermissionAdministrator) != 0) || (i.Member.Permissions&manage_audit) == manage_audit {
		return true
	}
	respond(s, i, "You do not have an administrator role! Permission Denied")
	return false
}
