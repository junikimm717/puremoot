package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/joho/godotenv/autoload"
)

var (
	dg *discordgo.Session
	db *Database
)

var (
	cleanup = flag.Bool("cleanup", false, "do we delete all application commands after the bot exits?")
)

func init() {
	flag.Parse()
	var err error
	dg, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}
}

func init() {
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		}
	})
	dg.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentGuilds | discordgo.IntentGuildMembers
}

func main() {
	err := dg.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	db = InitDatabase()
	log.Println("Adding commands...")
	commands_list := map[string]int{}
	for _, v := range commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", v)
		log.Println("Adding", v.Name)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		commands_list[v.Name] = 1
	}
	// remove any commands that are not part of our list.
	all_commands, err := dg.ApplicationCommands(dg.State.User.ID, "")
	if err != nil {
		log.Panicf("Cannot get Application Commands: %v", err)
	}
	for _, v := range all_commands {
		if commands_list[v.Name] != 1 {
			err := dg.ApplicationCommandDelete(dg.State.User.ID, "", v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
			log.Printf("Deleting extraneous command: %v", v.Name)
		}
	}
	// code for waiting for websocket to close.
	defer dg.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop
	db.client.Save(ctx).Err()
	if *cleanup {
		// remove all commands.
		commands, err := dg.ApplicationCommands(dg.State.User.ID, "")
		if err != nil {
			log.Panicf("Cannot get Application Commands: %v", err)
		}
		for _, v := range commands {
			err := dg.ApplicationCommandDelete(dg.State.User.ID, "", v.ID)
			if err != nil {
				log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
			}
		}
	}
	log.Println("Graceful Shutdown")
}
