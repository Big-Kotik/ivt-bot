package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Big-Kotik/ivt-pull-api/pkg/api"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	botToken := os.Getenv("TELEGRAM_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Println(err)
	}

	bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 1
	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			if update.Message.Document != nil {
				fileId := update.Message.Document.FileID
				log.Printf("[%s] %s", fileId, update.Message.Document.FileName)
				file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileId})
				if err != nil {
					log.Println(err)
				}

				fileLink := file.Link(botToken)
				log.Printf("%s", fileLink)

				requestsWrapper, err := parse(fileLink)
				if err != nil {
					log.Println(err)
				}

				log.Printf("RequestsWrapper: %s", requestsWrapper.Data)

				logPrintRequestsWrapper(requestsWrapper)

				conn, err := grpc.Dial("0.0.0.0:7272", grpc.WithInsecure()) // todo think about deprecated
				if err != nil {
					grpclog.Fatalf("fail to dial: %v", err)
				}

				response, err := resendRequest(requestsWrapper, conn)

				for {
					resp, err := response.Recv()
					if errors.Is(err, io.EOF) {
						break
					}
					if err != nil {
						log.Printf("fail to dial: %v", err)
						break
					}

					makeResponse(resp, update.Message, bot)
					log.Printf("Body: %s", string(resp.Body))
				}

				conn.Close()
			}
		}
	}
}

func logPrintRequestsWrapper(wrapper RequestsWrapper) { // optional function for understand what's going on
	for _, requestWrapper := range wrapper.Data {
		log.Printf("Next request: \n "+
			"Url: %s \n"+
			"Method: %s \n"+
			"Body: %s \n"+
			"Headers: %s \n", requestWrapper.Method, requestWrapper.Url, requestWrapper.Body, requestWrapper.Headers)
	}
}

func resendRequest(requestsWrapper RequestsWrapper,
	conn *grpc.ClientConn) (api.Puller_PullResourceClient, error) { // todo think about errors

	client := api.NewPullerClient(conn)
	request := createRequest(requestsWrapper)

	response, err := client.PullResource(context.Background(), request)
	if err != nil {
		log.Println(err)
	}

	return response, nil
}

func createRequest(requestsWrapper RequestsWrapper) *api.HttpRequestsWrapper {
	requests := make([]*api.HttpRequestsWrapper_Request, 0)
	for _, rd := range requestsWrapper.Data { // todo rename
		headers := make(map[string]*api.Header, 0)
		for key, value := range rd.Headers {
			headers[key] = &api.Header{Keys: value}
		}

		requests = append(requests, &api.HttpRequestsWrapper_Request{
			Url:     rd.Url,
			Body:    string(rd.Body), // TODO!!!! change to []byte in server
			Headers: headers,
			Method:  rd.Method,
		})
	}

	request := &api.HttpRequestsWrapper{
		Requests: requests,
	}
	return request
}

func makeResponse(response *api.Response, message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	chatId := message.Chat.ID
	userName := message.From.UserName
	fileName := userName + "_" + time.Now().Format("2006-01-02_15-04-05_") + "*.json"
	log.Printf("File name: %s", fileName)

	dirName, err := os.MkdirTemp("", userName) // create temp directory
	if err != nil {
		log.Fatal(err)
	}

	tempFile, err := os.CreateTemp(dirName, fileName) // create temp file
	if err != nil {
		log.Fatal(err)
	}

	content, err := json.Marshal(createResponse(response))
	if err != nil {
		fmt.Println(err)
		return
	}

	if _, err := tempFile.Write(content); err != nil {
		log.Fatal(err)
	}

	log.Printf("File name: %s, dir name: %s", tempFile.Name(), dirName)
	msg := tgbotapi.NewDocument(chatId, tgbotapi.FilePath(tempFile.Name()))
	_, err = bot.Send(msg)

	if err != nil {
		log.Fatal(err)
	}
	tempFile.Close() // I don't use defer because he don't delete files
	os.Remove(tempFile.Name())
	os.RemoveAll(dirName)
}

func createResponse(response *api.Response) *ResponseWrapper {
	headers := make(map[string][]string, 0)
	for key, value := range response.Header {
		headers[key] = value.Keys
	}
	responseWrapper := &ResponseWrapper{
		Body:    response.Body,
		Headers: headers,
	}

	return responseWrapper
}

func parse(link string) (RequestsWrapper, error) {
	var requests RequestsWrapper

	client := http.Client{}

	resp, err := client.Get(link)
	if err != nil {
		return requests, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&requests)
	if err != nil {
		return requests, err
	}

	return requests, nil
}
