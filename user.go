package main

import (
	"fmt"
	"github.com/garyburd/go-websocket/websocket"
	"github.com/ugorji/go/codec"
	"io/ioutil"
	"strconv"
	"time"
)

const (
	MAX_USERS_PER_LOCATION = 20
	MAX_NICKNAME_LENGTH    = 20
	SEND_CHANNEL_SIZE      = 256
	// Time allowed to write a message to the client.
	WRITE_WAIT = 5 * time.Second
	// Time allowed to read the next message from the client.
	READ_WAIT = 20 * time.Second
	// Send pings to client with this period. Must be less than READ_WAIT.
	PING_PERIOD = READ_WAIT / 2
	// Maximum message size allowed from client.
	// must be at least 4 * MAX_CHAT_MESSAGE_LENGTH (in script.js)
	MAX_MESSAGE_SIZE = 1024
)

var (
	userIdGenerator chan int
	msgpackHandle   codec.MsgpackHandle
)

type User struct {
	Socket      *websocket.Conn
	sendData    chan<- *[]byte
	recv        <-chan *[]interface{}
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

func NewUser(ws *websocket.Conn) *User {
	user := new(User)
	user.Socket = ws
	user.UserId = <-userIdGenerator
	user.Nickname = strconv.Itoa(user.UserId)
	user.UsePen = true
	user.sendData = user.sender()
	user.recv = user.receiver()
	user.kick = make(chan string)
	return user
}

func init() {
	userIdGenerator = make(chan int)
	go func() {
		i := 1
		for {
			userIdGenerator <- i
			i += 1
		}
	}()
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

func (user *User) Error(description string) {
	user.SendEvent(&[]interface{}{EventTypeError, description})
	user.Kick(description)
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
		Log.Printf("Buffer full for user %v: kicking.\n", user.UserId)
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

func (user *User) sender() chan<- *[]byte {
	ch := make(chan *[]byte, SEND_CHANNEL_SIZE)
	go func() {
		ticker := time.NewTicker(PING_PERIOD)
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			case data, ok := <-ch:
				if !ok {
					write(user.Socket, websocket.OpClose, []byte{})
					return
				}
				if err := write(user.Socket, websocket.OpBinary, *data); err != nil {
					user.Kick(err.Error())
					return
				}
			case <-ticker.C:
				if err := write(user.Socket, websocket.OpPing, []byte{}); err != nil {
					user.Kick(err.Error())
					return
				}
			}
		}
	}()
	return ch
}

func (user *User) receiver() <-chan *[]interface{} {
	// receives and decodes messages from users
	ch := make(chan *[]interface{})
	go func() {
		user.Socket.SetReadLimit(MAX_MESSAGE_SIZE)
		user.Socket.SetReadDeadline(time.Now().Add(READ_WAIT))
		for {
			op, r, err := user.Socket.NextReader()
			if err != nil {
				user.Kick(err.Error())
				break
			}
			switch op {
			case websocket.OpPong:
				user.Socket.SetReadDeadline(time.Now().Add(READ_WAIT))
			case websocket.OpBinary:
				data, err := ioutil.ReadAll(r)
				if err != nil {
					user.Kick(err.Error())
					break
				}
				var event []interface{}
				err = codec.NewDecoderBytes(data, &msgpackHandle).Decode(&event)
				if err != nil {
					user.Kick(err.Error())
					break
				}
				ch <- &event
			default:
				user.Kick(fmt.Sprintf("bad message type: %v", op))
				break
			}
		}
		close(ch)
	}()
	return ch
}

func (user *User) SocketHandler(location *Location) {
	defer func() {
		if location != nil {
			location.Quit <- user
			Log.Printf("user %v left\n", user.UserId)
		}
	}()
	for {
		select {
		case event, ok := <-user.recv:
			if !ok {
				return
			}
			if location != nil {
				location.Message <- &UserAndEvent{user, event}
			}
		case err_msg := <-user.kick:
			Log.Printf("user %v kicked for '%v'\n", user.UserId, err_msg)
			return
		}
	}
}
