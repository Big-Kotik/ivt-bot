# syntax=docker/dockerfile:1

FROM golang:1.19.1-alpine

WORKDIR /app

COPY go.mod go.sum ./
COPY ./ivt-pull-api ./
RUN go mod download
RUN go mod tidy

COPY . ./

RUN go build -o /ivt-bot ./cmd/proxy-bot

ENTRYPOINT [ "/ivt-bot" ]