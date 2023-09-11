package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

/*
CRUD for pureMOOt manager role
*/

func (d *Database) GetManager(guildId string) (string, bool) {
	return d.GetString(
		fmt.Sprintf("manager:%v", guildId),
	)
}

func (d *Database) SetManager(guildId string, roleId string) error {
	return d.SetString(
		fmt.Sprintf("manager:%v", guildId),
		roleId,
	)
}

func (d *Database) UnSetManager(guildId string) error {
	return d.client.Del(
		ctx,
		fmt.Sprintf("manager:%v", guildId),
	).Err()
}

func forceManager(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Member == nil {
		respond(s, i, "You cannot invoke this command outside of a guild")
		return false
	}
	manager, exists := db.GetManager(i.GuildID)
	if exists {
		for _, role := range i.Member.Roles {
			if role == manager {
				return true
			}
		}
	} else {
		manage_audit := int64(discordgo.PermissionManageServer | discordgo.PermissionViewAuditLogs)
		if (i.Member.Permissions&(discordgo.PermissionAdministrator) != 0) || (i.Member.Permissions&manage_audit) == manage_audit {
			return true
		}
	}
	respond(s, i, "You do not have the pureMOOt manager role! Permission Denied")
	return false
}

var ManagerHandlers = map[string]SubcommandHandler{
	"get": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		manager, exists := db.GetManager(i.GuildID)
		if !exists {
			respond(s, i, "No manager role exists for pureMOOt!")
		} else {
			respond(s, i, fmt.Sprintf("The <@&%v> role can manage pureMOOt", manager))
		}
	},
	"set": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		options := extractInteractionOptions(opts)
		role := options["role"].RoleValue(s, i.GuildID)
		err := db.SetManager(i.GuildID, role.ID)
		if err != nil {
			respond(
				s, i,
				fmt.Sprintf("Error! %v", err.Error()),
			)
		} else {
			respond(
				s,
				i,
				fmt.Sprintf("Manager Role set to %v", role.Mention()),
			)
		}
	},
	"unset": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		err := db.UnSetManager(i.GuildID)
		if err != nil {
			respond(
				s, i,
				fmt.Sprintf("Error! %v", err.Error()),
			)
		} else {
			respond(
				s,
				i,
				"Manager role has been unset",
			)
		}
	},
	"bban": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		err := db.DisableBroadcast(i.ChannelID)
		if err != nil {
			respond(
				s, i,
				fmt.Sprintf("Error! %v", err.Error()),
			)
		} else {
			respond(
				s,
				i,
				"Broadcast is now disabled",
			)
		}
	},
	"ballow": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		err := db.EnableBroadcast(i.ChannelID)
		if err != nil {
			respond(
				s, i,
				fmt.Sprintf("Error! %v", err.Error()),
			)
		} else {
			respond(
				s,
				i,
				"Broadcast is now allowed",
			)
		}
	},
	"bpermitted": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if db.BroadcastAllowed(i.ChannelID) {
			respond(s, i, "Broadcast is permitted on this channel!")
		} else {
			respond(s, i, "Broadcast is banned on this channel!")
		}
	},
}
