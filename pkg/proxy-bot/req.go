package proxybot

import "github.com/google/uuid"

type RequestWrapper struct {
	Uuid    uuid.UUID           `json:"id"`
	Url     string              `json:"url"`
	Method  string              `json:"method"`
	Body    []byte              `json:"body"`
	Headers map[string][]string `json:"headers"`
}

type RequestsWrapper struct {
	Data []RequestWrapper `json:"data"`
}

type ResponseWrapper struct {
	Uuid          uuid.UUID           `json:"id"`
	StatusCode    int32               `json:"code"`
	ProtoMajor    int32               `json:"protoMajor"`
	ProtoMinor    int32               `json:"protoMinor"`
	Header        map[string][]string `json:"header"`
	Body          []byte              `json:"body"`
	ContentLength int64               `json:"contentLength"`
}

type ResponsesWrapper struct {
	Data []ResponseWrapper `json:"data"`
}
