package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/ugorji/go/codec"
	"io/ioutil"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	MAX_USERS_PER_LOCATION = 20
	MAX_NICKNAME_LENGTH    = 20
	SEND_CHANNEL_SIZE      = 256
	// Time allowed to write a message to the peer.
	WRITE_WAIT = 5 * time.Second
	// Time allowed to read the next pong message from the peer.
	PONG_WAIT = 20 * time.Second
	// Send pings to client with this period. Must be less than PONG_WAIT.
	PING_PERIOD = PONG_WAIT / 2
	// Maximum message size allowed from peer.
	// must be at least 4 * MAX_CHAT_MESSAGE_LENGTH (in script.js)
	MAX_MESSAGE_SIZE      = 1024
	LAST_USER_ID_FILENAME = "last_user_id"
)

var (
	lastUserId    uint32
	msgpackHandle codec.MsgpackHandle
)

type User struct {
	sendData    chan *[]byte
	UserId      int
	Nickname    string
	MouseIsDown bool
	PositionX   int
	PositionY   int
	ColorRed    int
	ColorGreen  int
	ColorBlue   int
	UsePen      bool
	kick        chan string
}

func NewUser() *User {
	user := new(User)
	user.UserId = int(atomic.AddUint32(&lastUserId, 1))
	user.Nickname = strconv.Itoa(user.UserId)
	user.UsePen = true
	user.sendData = make(chan *[]byte, SEND_CHANNEL_SIZE)
	user.kick = make(chan string)
	return user
}

func init() {
	lastUserId = ReadIntFromFile(LAST_USER_ID_FILENAME)
}

func encodeEvent(event *[]interface{}) (*[]byte, error) {
	var result []byte
	var err = codec.NewEncoderBytes(&result, &msgpackHandle).Encode(event)
	if err != nil {
		Log.Printf("Failed to encode event %v\n", event)
	}
	return &result, err
}

func (user *User) Kick(description string) {
	select {
	case user.kick <- description:
	default:
		// SocketHandler has already returned, avoid blocking
	}
}

func (user *User) SendEvent(event *[]interface{}) {
	data, err := encodeEvent(event)
	if err != nil {
		return
	}
	user.SendData(data)
}

func (user *User) SendData(data *[]byte) {
	select {
	case user.sendData <- data:
	default:
		user.Kick("Buffer full")
	}
}

func (user *User) mouseMove(x int, y int, duration int) {
	user.PositionX = x
	user.PositionY = y
}

func (user *User) mouseUp() {
	user.MouseIsDown = false
}

func (user *User) mouseDown() {
	user.MouseIsDown = true
}

func (user *User) changeTool(use_pen bool) {
	user.UsePen = use_pen
}

func (user *User) changeColor(red, green, blue int) {
	user.ColorRed = red
	user.ColorGreen = green
	user.ColorBlue = blue
}

func (user *User) changeNickname(nickname string) {
	user.Nickname = nickname
}

func write(ws *websocket.Conn, opCode int, payload []byte) error {
	ws.SetWriteDeadline(time.Now().Add(WRITE_WAIT))
	return ws.WriteMessage(opCode, payload)
}

func (user *User) Sender(ws *websocket.Conn) {
	go func() {
		ticker := time.NewTicker(PING_PERIOD)
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			case data, open := <-user.sendData:
				if !open {
					// channel is closed: user left
					return
				}
				if err := write(ws, websocket.BinaryMessage, *data); err != nil {
					user.Kick(err.Error())
					return
				}
			case <-ticker.C:
				if err := write(ws, websocket.PingMessage, []byte{}); err != nil {
					user.Kick(err.Error())
					return
				}
			}

		}
	}()
}

func (user *User) Receiver(location *Location, ws *websocket.Conn) {
	// receives and decodes messages from users
	go func() {
		ws.SetReadLimit(MAX_MESSAGE_SIZE)
		ws.SetReadDeadline(time.Now().Add(PONG_WAIT))
		ws.SetPongHandler(func(string) error {
			ws.SetReadDeadline(time.Now().Add(PONG_WAIT))
			return nil
		})
		for {
			op, r, err := ws.NextReader()
			if err != nil {
				user.Kick(err.Error())
				return
			}
			switch op {
			case websocket.BinaryMessage:
				data, err := ioutil.ReadAll(r)
				if err != nil {
					user.Kick(err.Error())
					return
				}
				var event []interface{}
				err = codec.NewDecoderBytes(data, &msgpackHandle).Decode(&event)
				if err != nil {
					user.Kick(err.Error())
					return
				}
				location.Message <- &UserAndEvent{user, &event}
			default:
				user.Kick(fmt.Sprintf("bad message type: %v", op))
				return
			}
		}
	}()
}
