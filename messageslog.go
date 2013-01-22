package main

import (
	"encoding/gob"
	"os"
)

const (
	MAX_MESSAGES_PER_CHAT = 100
)

type Message struct {
	Nickname string
	Message string
}

type MessagesLog struct {
	Messages []*Message
	filename string
	dirty bool
}

func OpenMessagesLog(filename string) *MessagesLog {
	result, err := openMessagesLog(filename)
	if err != nil {
		result = new(MessagesLog)
	}
	result.filename = filename
	result.dirty = false
	return result
}

func (chat *MessagesLog) AddMessage(nickname string, msg string) {
	chat.dirty = true
	if len(chat.Messages) == MAX_MESSAGES_PER_CHAT {
		chat.Messages = chat.Messages[1:]
	}
	message := new(Message)
	message.Nickname = nickname
	message.Message = msg
	chat.Messages = append(chat.Messages, message)
}

func (chat *MessagesLog) GetMessages() (result [][2]string) {
	for _, message := range chat.Messages {
		result = append(result, [2]string{message.Nickname, message.Message})
	}
	return result
}

func (chat *MessagesLog) Save() {
	if chat.dirty {
		chat.serialize()
	}
}

func (chat *MessagesLog) serialize() {
	file_handle, err := os.OpenFile(chat.filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		Log.Println("Couldn't open", chat.filename, "for writing")
	}
	defer file_handle.Close()
	gob.NewEncoder(file_handle).Encode(chat)
	chat.dirty = false
}

func openMessagesLog(filename string) (*MessagesLog, error) {
	file_handle, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file_handle.Close()
	var chat *MessagesLog
	decoder := gob.NewDecoder(file_handle)
	err = decoder.Decode(&chat)
	return chat, err
}
