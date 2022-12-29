package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/otiai10/openaigo"
)

var gptClient *openaigo.Client

func main() {

	// Get Tokens
	gptToken := os.Getenv("GPT_TOKEN")
	if gptToken == "" {
		panic("GPT_TOKEN env variable must be set")
	}
	discordToken := os.Getenv("DISCORD_CLIENT_TOKEN")
	if discordToken == "" {
		panic("DISCORD_CLIENT_TOKEN env variable must be set")
	}

	// Init clients
	gptClient = openaigo.NewClient(gptToken)
	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		panic(err)
	}

	// Handle discord messages
	dg.AddHandler(SearchHandler)
	dg.Identify.Intents = discordgo.IntentGuildMessages

	// Open Bot
	err = dg.Open()
	if err != nil {
		panic(err)
	}

	// Wait close
	fmt.Println("Bot is now running. Press CTRL-C to exit...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Clearly close
	dg.Close()
}

func SearchHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Split message in string array
	splittedContent := strings.Split(m.Content, " ")

	// If there is not arguments, return
	if len(splittedContent) <= 1 {
		return
	}

	// If the message is the GPT command
	if splittedContent[0] == "!gpt" {
		// Get command arguments string
		args := strings.Join(splittedContent[1:], " ")

		// Request to ChatGPT
		resp, err := gptClient.Completion(nil, openaigo.CompletionRequestBody{
			Model:  "text-davinci-003",
			Prompt: []string{args},
		})
		if err != nil {
			s.ChannelMessageSendReply(m.ChannelID, "Oups! An error occured...", m.Reference())
			return
		}
		s.ChannelMessageSendReply(m.ChannelID, resp.Choices[0].Text, m.Reference())
	}
}
