package main

import "C"
import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sashabaranov/go-openai"
)

type Config struct {
	DiscordSessionToken string `env:"DISCORD_SESSION_TOKEN" flag:"sessionToken" flagDesc:"The Session token for the Discord Bot"`
	APIToken            string `env:"API_TOKEN" flag:"apitoken" flagDesc:"The api token"`
}

type DBot struct {
	Discord *discordgo.Session
	Config  Config
	GPT     *openai.Client
}

func (b *DBot) Connect() error {
	var err error

	clientConfig := openai.DefaultConfig(b.Config.APIToken)

	clientConfig.BaseURL = "http://ollama:11434/v1/"

	b.GPT = openai.NewClientWithConfig(clientConfig)

	b.Discord, err = discordgo.New("Bot " + b.Config.DiscordSessionToken)
	if err != nil {
		fmt.Println("error opening connection,", err)
	}
	// Handlers
	b.Discord.AddHandler(b.RespondGPT)
	b.Discord.Identify.Intents = discordgo.IntentsGuildMessages

	err = b.Discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
	}

	return err
}

func (b *DBot) RespondGPT(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if !containsUser(m.Content, "<@"+s.State.User.ID+">") {
		return
	}
	fmt.Println("Responding...")

	lastMessages, _ := s.ChannelMessages(m.ChannelID, 5, "", "", "")

	var gptMessages []openai.ChatCompletionMessage

	var generatedHistory string

	for _, element := range lastMessages {
		generatedHistory = generatedHistory + element.Author.Username + ": " + element.Content + "\n"
	}

	gptMessages = append(gptMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: generatedHistory,
	})

	resp, err := b.GPT.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    "samewise",
			Messages: gptMessages,
		},
	)

	fmt.Println("Logging Response:")
	fmt.Println(resp)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	s.ChannelMessageSend(m.ChannelID, resp.Choices[0].Message.Content)
}

func containsUser(s string, user string) bool {
	return strings.Contains(s, user)
}
