package main

import (
	"flag"
	"fmt"
	"os"

	proxybot "github.com/Big-Kotik/ivt-bot/pkg/proxy-bot"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var serverAddr = flag.String("server_addr", "localhost:8080", "The server address in the format of host:port")

func main() {

	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
