package main

const (
	MAX_MESSAGES_PER_CHAT = 100
)

type Message struct {
	Nickname string
	Message string
}

type MessagesLog struct {
	Messages []*Message
}

func NewMessagesLog() *MessagesLog {
	chat := new(MessagesLog)
	return chat
}

func (chat *MessagesLog) AddMessage(nickname string, msg string) {
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


