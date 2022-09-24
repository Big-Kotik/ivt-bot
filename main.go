package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Big-Kotik/ivt-pull-api/pkg/api"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	//runServer()
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
			log.Printf("[%s]", update.Message.From.UserName)
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

				for _, requestWrapper := range requestsWrapper.Data {
					log.Printf("Next request: \n Url: %s \n Method: %s \n Body: %s \n Headers: %s \n ", requestWrapper.Method, requestWrapper.Url, requestWrapper.Body, requestWrapper.Headers)

				}
				resendRequest(requestsWrapper)
			}
		}
	}
}

func runServer() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Create a gRPC server object
	s := grpc.NewServer()

	go func() {
		log.Fatalln(s.Serve(lis))
	}()
}

func resendRequest(requestsWrapper RequestsWrapper) error {
	conn, err := grpc.Dial("0.0.0.0:8080", grpc.WithInsecure())
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := api.NewPullerClient(conn)

	requests := make([]*api.HttpRequestsWrapper_Request, 0)
	for _, rd := range requestsWrapper.Data { // todo rename
		headers := make(map[string]*api.Header, 0)
		for key, value := range rd.Headers {
			headers[key] = &api.Header{Keys: value}
		}

		requests = append(requests, &api.HttpRequestsWrapper_Request{
			Url:     rd.Url,
			Body:    rd.Body,
			Headers: headers,
			Method:  rd.Method,
		})
	}

	request := &api.HttpRequestsWrapper{
		Requests: requests,
	}

	response, err := client.PullResource(context.Background(), request)

	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}

	fmt.Println(response.Context())
	return nil
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
