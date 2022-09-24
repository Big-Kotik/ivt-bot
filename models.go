package main

type RequestWrapper struct {
	Url     string              `json:"url"`
	Method  string              `json:"method"`
	Body    string              `json:"body"`
	Headers map[string][]string `json:"headers"`
}

type RequestsWrapper struct {
	Data []RequestWrapper `json:"data"`
}
