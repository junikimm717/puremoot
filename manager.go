package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

/*
CRUD for pureMOOt manager,cow role
*/

const (
	PuremootManagerRole = "manager"
	PuremootCowRole     = "cow"
)

var PuremootRoleOptions = discordgo.ApplicationCommandOption{
	Type:        discordgo.ApplicationCommandOptionString,
	Name:        "roletype",
	Description: "Role Type",
	Choices: []*discordgo.ApplicationCommandOptionChoice{
		{
			Name:  "Manager",
			Value: PuremootManagerRole,
		},
		{
			Name:  "Cow",
			Value: PuremootCowRole,
		},
	},
	Required: true,
}

func (d *Database) GetPuremootRole(puremootRole string, guildId string) (string, bool) {
	return d.GetString(
		fmt.Sprintf("%v:%v", puremootRole, guildId),
	)
}

func (d *Database) SetPuremootRole(puremootRole string, guildId string, roleId string) error {
	return d.SetString(
		fmt.Sprintf("%v:%v", puremootRole, guildId),
		roleId,
	)
}

func (d *Database) UnsetPuremootRole(puremootRole string, guildId string) error {
	return d.client.Del(
		d.ctx,
		fmt.Sprintf("%v:%v", puremootRole, guildId),
	).Err()
}

func forceManager(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	if i.Member == nil {
		respond(s, i, "You cannot invoke this command outside of a guild")
		return false
	}
	manager, exists := db.GetPuremootRole(PuremootManagerRole, i.GuildID)
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
	"rban": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		options := extractInteractionOptions(opts)
		user := options["user"].UserValue(s)
		db.BanReaperUser(user.ID, i.ChannelID)
		respond(s, i, fmt.Sprintf("User %v has been banned from reaping on this channel!", user.Username))
	},
	"rallow": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		options := extractInteractionOptions(opts)
		user := options["user"].UserValue(s)
		db.AllowReaperUser(user.ID, i.ChannelID)
		respond(s, i, fmt.Sprintf("User %v has been unbanned from reaping on this channel!", user.Username))
	},
	"rpermitted": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		options := extractInteractionOptions(opts)
		user := options["user"].UserValue(s)
		if db.IsReaperUserBanned(user.ID, i.ChannelID) {
			respond(s, i, fmt.Sprintf("User %v is banned from reaping on this channel!", user.Username))
		} else {
			respond(s, i, fmt.Sprintf("User %v is allowed to reap on this channel!", user.Username))
		}
	},
	"get": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		options := extractInteractionOptions(opts)
		puremootRole := options["roletype"].StringValue()
		roleId, exists := db.GetPuremootRole(puremootRole, i.GuildID)
		if !exists {
			respond(s, i, fmt.Sprintf("No %v role exists for pureMOOt!", puremootRole))
		} else {
			respond(s, i, fmt.Sprintf("The <@&%v> Discord role is the %v role for pureMOOt", roleId, puremootRole))
		}
	},
	"set": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		options := extractInteractionOptions(opts)
		puremootRole := options["roletype"].StringValue()
		role := options["role"].RoleValue(s, i.GuildID)
		err := db.SetPuremootRole(puremootRole, i.GuildID, role.ID)
		if err != nil {
			respond(
				s, i,
				fmt.Sprintf("Error! %v", err.Error()),
			)
		} else {
			respond(
				s,
				i,
				fmt.Sprintf("puremoot role %v set to %v", puremootRole, role.Mention()),
			)
		}
	},
	"unset": func(s *discordgo.Session, i *discordgo.InteractionCreate, opts []*discordgo.ApplicationCommandInteractionDataOption) {
		if !forceManager(s, i) {
			return
		}
		options := extractInteractionOptions(opts)
		puremootRole := options["roletype"].StringValue()
		err := db.UnsetPuremootRole(puremootRole, i.GuildID)
		if err != nil {
			respond(
				s, i,
				fmt.Sprintf("Error! %v", err.Error()),
			)
		} else {
			respond(
				s,
				i,
				fmt.Sprintf("%v Role unset", puremootRole),
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
