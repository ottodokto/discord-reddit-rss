package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mmcdole/gofeed"
)

var (
	botToken   = os.Getenv("DISCORD_BOT_TOKEN")
	channelID  = os.Getenv("DISCORD_CHANNEL_ID")
	rssURL     = os.Getenv("RSS_FEED_URL")
	searchTerm string
	lastPost   = time.Now() // Track the last time we checked the feed
)

func main() {
	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("error creating Discord session: %v", err)
	}

	dg.AddMessageCreateHandler(messageHandler)

	if err := dg.Open(); err != nil {
		log.Fatalf("error opening Discord connection: %v", err)
	}
	defer dg.Close()

	go func() {
		for {
			checkFeed(dg)
			time.Sleep(10 * time.Minute)
		}
	}()

	fmt.Println("Bot is running...")
	select {}
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || !strings.HasPrefix(m.Content, "!setsearch") {
		return
	}

	// Set search term
	searchTerm = strings.TrimSpace(m.Content[len("!setsearch "):])
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Search term updated to: %s", searchTerm))
}

func checkFeed(s *discordgo.Session) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(rssURL, context.Background())
	if err != nil {
		log.Printf("error fetching RSS feed: %v", err)
		return
	}

	for _, item := range feed.Items {
		if item.PublishedParsed != nil && item.PublishedParsed.After(lastPost) && searchTerm != "" {
			if strings.Contains(strings.ToLower(item.Title), strings.ToLower(searchTerm)) ||
				strings.Contains(strings.ToLower(item.Description), strings.ToLower(searchTerm)) {
				_, _ = s.ChannelMessageSend(channelID, fmt.Sprintf("New post matching '%s': %s", searchTerm, item.Link))
			}
		}
	}
	lastPost = time.Now() // Update last checked time
}
