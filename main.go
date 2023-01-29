package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

		// Response text message
		response := make(chan string)

		// Get command arguments string
		args := strings.Join(splittedContent[1:], " ")

		// Send message before processing GPT search
		message, _ := s.ChannelMessageSendReply(m.ChannelID, "Processing...", m.Reference())

		// Handle timeouts
		go GetGptResponse(args, response)
		select {
		case text := <-response:
			// Edit sent message with GPT response or GPT error response
			s.ChannelMessageEdit(m.ChannelID, message.ID, text)
		case <-time.After(1 * time.Minute):
			// Edit send message with timeout error
			s.ChannelMessageEdit(m.ChannelID, message.ID, "Désolé chakal chu ko là mon reuf")
		}
	}
}

func GetGptResponse(args string, response chan string) {

	// Request to ChatGPT
	gptResp, err := gptClient.Completion(context.Background(), openaigo.CompletionRequestBody{
		Model:     "text-davinci-003",
		Prompt:    []string{args},
		MaxTokens: 2048, // Max for this model
	})
	if err != nil || len(gptResp.Choices) < 1 {
		response <- "Oups! An error occured..."
	} else {
		// Set default response message by an error
		response <- gptResp.Choices[0].Text
	}
}
