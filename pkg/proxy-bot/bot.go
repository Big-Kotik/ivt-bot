package proxybot

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Big-Kotik/ivt-pull-api/pkg/api"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type ProxyBot struct {
	*tgbotapi.BotAPI
	puller *grpc.ClientConn
	logger log.Logger
}

func NewProxyBot(token string, puller *grpc.ClientConn) (*ProxyBot, error) {
	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		return nil, err
	}

	bot.Debug = true

	return &ProxyBot{bot, puller, *log.Default()}, nil
}

func (b *ProxyBot) Listen() {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 1

	updates := b.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil || update.Message.Document == nil {
			continue
		}

		// TODO: check file size
		fileUrl, err := b.GetFileDirectURL(update.Message.Document.FileID)

		if err != nil {
			b.logger.Printf("can't get file url - %s", err.Error())
			continue
		}

		requests, err := getRequests(fileUrl)

		if err != nil {
			b.logger.Printf("can't get requests - %s", err.Error())
			continue
		}

		// TODO: хранить сразу его?
		puller := api.NewPullerClient(b.puller)
		stream, err := puller.PullResources(context.Background(), newGrpcRequests(requests))

		fmt.Printf("------------ pull all --------")

		// TODO: send error
		if err != nil {
			b.logger.Printf("can't pullResources - %s", err.Error())
			fmt.Printf("---------- errors ocurred ---------")
			continue
		}

		i := 0

		for {
			resp, err := stream.Recv()
			i++
			if err == io.EOF {
				fmt.Println(i)
				break
			}

			if err != nil {
				fmt.Printf("err != nil: %s", err.Error())
				break
			}

			err = b.sendResponse(update.Message.Chat.ID, resp)

			if err != nil {
				b.logger.Printf("cant' send response - %s", err.Error())
			}
		}
	}
}

func getRequests(url string) (*RequestsWrapper, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	fmt.Println("----------------------\nstart get req ----------------------")
	defer resp.Body.Close()

	var requests RequestsWrapper

	err = json.NewDecoder(resp.Body).Decode(&requests)

	return &requests, err
}

// rs is not nil
func newGrpcRequests(rs *RequestsWrapper) *api.HttpRequests {
	grpcRequests := api.HttpRequests{Requests: make([]*api.HttpRequests_HttpRequest, 0)}

	// TODO: typeof(HttoReq...) == string??!!
	for _, req := range rs.Data {
		headers := make(map[string]*api.Header)

		for k, h := range req.Headers {
			headers[k] = &api.Header{Keys: h}
		}

		grpcRequests.Requests = append(grpcRequests.Requests, &api.HttpRequests_HttpRequest{
			Url:     req.Url,
			Method:  req.Method,
			Body:    string(req.Body),
			Headers: headers,
		})
	}

	return &grpcRequests
}

func (b *ProxyBot) sendResponse(chatId int64, resp *api.HttpResponse) error {
	fmt.Println("================\n start sending response \n==========")
	fileName := strconv.FormatInt(chatId, 10) + "_" + time.Now().Format("2006-01-02_15-04-05_") + "*.json"
	log.Printf("File name: %s", fileName)

	dirName, err := os.MkdirTemp("", strconv.FormatInt(chatId, 10)) // create temp directory
	if err != nil {
		log.Fatal(err)
	}

	tempFile, err := os.CreateTemp(dirName, fileName) // create temp file
	if err != nil {
		log.Fatal(err)
	}

	headers := make(map[string][]string)

	for k, v := range resp.Header {
		headers[k] = v.Keys
	}

	uid, _ := uuid.FromBytes(resp.Uuid)

	content, err := json.Marshal(ResponseWrapper{
		StatusCode:    resp.StatusCode,
		ProtoMajor:    resp.ProtoMajor,
		ProtoMinor:    resp.ProtoMinor,
		Header:        headers,
		Body:          resp.Body,
		ContentLength: resp.ContentLength,
		Uuid:          uid,
	})

	if err != nil {
		b.logger.Printf("can't marshal response - %s", err.Error())
		return err
	}

	if _, err := tempFile.Write(content); err != nil {
		log.Fatal(err)
	}

	log.Printf("File name: %s, dir name: %s", tempFile.Name(), dirName)
	msg := tgbotapi.NewDocument(chatId, tgbotapi.FilePath(tempFile.Name()))
	_, err = b.Send(msg)

	if err != nil {
		log.Fatal(err)
	}
	tempFile.Close() // I don't use defer because he don't delete files
	os.Remove(tempFile.Name())
	os.RemoveAll(dirName)

	return nil
}
