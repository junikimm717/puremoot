package main

import (
	"errors"
	"math/rand"
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

func cowRole(s *discordgo.Session, i *discordgo.InteractionCreate) (string, error) {
	roleId, exists := db.GetPuremootRole(PuremootCowRole, i.GuildID)
	if exists {
		return roleId, nil
	}
	roles, err := s.GuildRoles(i.GuildID)
	if err != nil {
		return "", err
	}
	for _, role := range roles {
		if strings.ToLower(role.Name) == "cow" {
			return role.ID, nil
		}
	}
	return "", errors.New("no server role 'cow'")
}

func getCows(s *discordgo.Session, i *discordgo.InteractionCreate) ([]*discordgo.Member, error) {
	cow_role_id, err := cowRole(s, i)
	if err != nil {
		return []*discordgo.Member{}, err
	}
	members, err := s.GuildMembers(i.GuildID, "", 800)
	if err != nil {
		return []*discordgo.Member{}, err
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
