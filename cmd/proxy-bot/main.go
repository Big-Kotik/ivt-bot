package main

import (
	"fmt"
	proxybot "ivt-bot/pkg/proxy-bot"
	"os"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(os.Getenv("grpc"), grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return
	}
	bot, err := proxybot.NewProxyBot(os.Getenv("token"), conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	bot.Listen()
}
