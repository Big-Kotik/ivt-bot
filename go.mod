module github.com/Big-Kotik/ivt-bot

go 1.18

require (
	github.com/Big-Kotik/ivt-pull-api v0.0.0-00010101000000-000000000000
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
	google.golang.org/grpc v1.49.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0
	golang.org/x/net v0.0.0-20220923203811-8be639271d50 // indirect
	golang.org/x/sys v0.0.0-20220919091848-fb04ddd9f9c8 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20220923205249-dd2d53f1fffc // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/Big-Kotik/ivt-pull-api => ./ivt-pull-api
