package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	botToken := os.Getenv("TELEGRAM_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 1
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s]", update.Message.From.UserName)
			if update.Message.Document != nil {
				fileId := update.Message.Document.FileID
				log.Printf("[%s] %s", fileId, update.Message.Document.FileName)
				file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileId})
				if err != nil {
					log.Panic(err)
				}

				fileLink := file.Link(botToken)
				log.Printf("%s", fileLink)

				savedFileName, err := DownloadFile(fileLink)
				if err != nil {
					panic(err)
				}

				log.Printf("Saved file name: %s", savedFileName)
			}
		}
	}
}

func DownloadFile(link string) (string, error) {
	fileURL, err := url.Parse(link)
	if err != nil {
		log.Fatal(err)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	file, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	client := http.Client{}

	resp, err := client.Get(link)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	size, err := io.Copy(file, resp.Body)

	defer file.Close()

	log.Printf("Downloaded a file %s with size %d", fileName, size)
	return fileName, nil
}
