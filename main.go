package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/saihon/saihon"
)

func main() {

	godotenv.Load()

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatalf("expected DISCORD_BOT_TOKEN in env variables")
	}

	chanID := os.Getenv("DISCORD_CHANID")
	if chanID == "" {
		log.Fatalf("expected DISCORD_CHANID in env variables")
	}

	url := os.Getenv("CRAWL_URL")
	if url == "" {
		log.Fatalf("expected CRAWL_URL in env variables")
	}

	notifyUser := os.Getenv("NOTIFY_USER")
	if notifyUser == "" {
		log.Fatalf("expected NOTIFY_USER in env variables")
	}

	client := &http.Client{
		Timeout: 4 * time.Second,
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}
	defer dg.Close()
	defer dg.ChannelMessageSend(chanID, "signed out")

	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)
	if err := dg.Open(); err != nil {
		log.Fatalf("FATAL: error opening connection: %v", err)
	}

	dg.ChannelMessageSend(chanID, "signed in")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	for {
		select {
		case <-sc:
			return
		case <-time.After(5 * time.Second):
			check(dg, client, url, chanID, notifyUser)
		}
	}
}

func check(dg *discordgo.Session, client *http.Client, url, chanID, notifyUser string) {

	log.Println("Checking komplett for graphics cards...")
	body, err := getResponse(client, url)
	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	hit, err := foundBuyButton(body)
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
	if hit {
		log.Println("for sale button found! ")
		if _, err := dg.ChannelMessageSend(chanID, fmt.Sprintf("new 3080 for sale <@%s>: %s", notifyUser, url)); err != nil {
			log.Printf("ERROR: %v", err)
		}
	}
}

func getResponse(client *http.Client, url string) ([]byte, error) {

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func foundBuyButton(body []byte) (bool, error) {

	document, err := saihon.Parse(strings.NewReader(string(body)))
	if err != nil {
		return false, err
	}

	result := document.Body().QuerySelectorAll(".product-list-item .buy-button")
	for element := range result.Enumerator() {
		contents := cleanContent(element.TextContent())
		return strings.Contains(contents, "köp"), nil
	}

	return false, nil
}

func cleanContent(input string) string {
	input = strings.ToLower(input)
	input = strings.Replace(input, "\n", "", -1)
	input = strings.Replace(input, "\t", "", -1)
	input = strings.Replace(input, " ", "", -1)
	input = strings.Replace(input, "-", "", -1)
	input = strings.Replace(input, "+", "", -1)
	input = strings.Replace(input, "tillagd", "", -1)
	return input
}
