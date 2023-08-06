package models

type Message struct {
	Client  Client
	Content string
	Time    int64
}
