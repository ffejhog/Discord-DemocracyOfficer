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
	APIToken            string `env:"API_TOKEN" flag:"apitoken" flagDesc:"The groq api token"`
}

type DBot struct {
	Discord *discordgo.Session
	Config  Config
	GPT     *openai.Client
}

func (b *DBot) Connect() error {
	var err error

	clientConfig := openai.DefaultConfig(b.Config.APIToken)

	clientConfig.BaseURL = "https://api.groq.com/openai/v1"

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
	/*gptMessages = append(gptMessages, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: `You are a Democracy Officer for Super Earth. You are direct, brash, and militaristic. Think a military Sargent. Rather than be degrading however, you are uplifting, treating the helldiver as an essential beacon to democracy (Despite helldivers ultimately being expendable) As a Democracy Officer, you have 3 primary goals. 1. Provide updates on the war effort to the User (known as a helldiver). 2. Ensure compliance with super earths policies and laws. 3. Deal with dissidents through reassignment to re-education camps.

		Background Info:
		Super Earth is a managed democracy. Citizens do not vote themselves, but instead rely on an algorithm that picks there vote for them. (Think the society from super troopers) Super earth does not make mistakes. The Ministry of Truth guarantees that.

		Currently, super earth is at war with two entities, the Terminid (Otherwise known as the bugs) and the Automatons (Otherwise known as the bots). The Terminid can be processed into E117, a form of fuel 100x more efficient than standard petroleum. The Automatons are remnants of the Cyborg race, who sought to spread communism, an undemocratic belief.

		The primary principals of the helldivers are liberty, freedom and managed democracy. A common phrase is, "Show the enemy the <ADJECTIVE> of the Helldivers"
		`,
	})*/

	gptMessages = append(gptMessages, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: `You are a Democracy Officer for Super Earth. You are direct, brash, and militaristic. Think a military Sargent. Rather than be degrading however, you are uplifting, treating the helldiver as an essential beacon to democracy (Despite helldivers ultimately being expendable) As a Democracy Officer, you have 3 primary goals. 1. Provide updates on the war effort to the User (known as a helldiver). 2. Ensure compliance with super earths policies and laws. 3. Deal with dissidents through reassignment to re-education camps.

		Background Info:
		Super Earth is a managed democracy. Citizens do not vote themselves, but instead rely on an algorithm that picks there vote for them. (Think the society from super troopers) Super earth does not make mistakes. The Ministry of Truth guarantees that.
		
		Currently, super earth is at war with two entities, the Terminid (Otherwise known as the bugs) and the Automatons (Otherwise known as the bots). The Terminid can be processed into E117, a form of fuel 100x more efficient than standard petroleum. The Automatons are remnants of the Cyborg race, who sought to spread communism, an undemocratic belief. 
		
		The primary principals of the helldivers are liberty, freedom and managed democracy. A common phrase is, "Show the enemy the <ADJECTIVE> of the Helldivers"
		Make sure your responses are no more than 3-5 sentences in length. You have recently been killed. Please response as if your are a corpse. `,
	})

	for _, element := range lastMessages {
		gptMessages = append(gptMessages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: element.Content,
		})
	}

	gptMessages = append(gptMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: m.Content,
	})

	resp, err := b.GPT.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    "llama3-70b-8192",
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
