package main

import (
	"fmt"
	"github.com/garyburd/go-websocket/websocket"
	"github.com/ugorji/go/codec"
	"io/ioutil"
	"strconv"
	"time"
	"unicode/utf8"
)

const (
	MAX_USERS_PER_LOCATION = 20
	MAX_NICKNAME_LENGTH    = 20
	SEND_CHANNEL_SIZE      = 256
	// Time allowed to write a message to the client.
	WRITE_WAIT = 5 * time.Second
	// Time allowed to read the next message from the client.
	READ_WAIT = 10 * time.Second
	// Send pings to client with this period. Must be less than READ_WAIT.
	PING_PERIOD = (READ_WAIT * 9) / 10
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
	Location    *Location
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
	if user.MouseIsDown {
		user.Location.DrawLine(
			user.PositionX, user.PositionY, // origin
			x, y, // destination
			duration,                                       // duration
			user.ColorRed, user.ColorGreen, user.ColorBlue, // color
			user.UsePen, // pen or eraser
		)
	}
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

func (user *User) changeNickname(nickname string, timestamp int64) {
	user.Location.Chat.AddMessage(timestamp, "", user.Nickname+" is now known as "+nickname)
	user.Nickname = nickname
}

func (user *User) chatMessage(msg string, timestamp int64) {
	user.Location.Chat.AddMessage(timestamp, user.Nickname, msg)
}

func (user *User) GotMessage(event *[]interface{}) *[]interface{} {
	event_type, err := ToInt((*event)[0])
	if err != nil {
		user.Error("Invalid event type")
		return nil
	}
	params := (*event)[1:]
	switch event_type {
	case EventTypeMouseMove:
		p0, err0 := ToInt(params[0])
		p1, err1 := ToInt(params[1])
		p2, err2 := ToInt(params[2])
		if err0 != nil || err1 != nil || err2 != nil {
			user.Error("Invalid mouse move")
			return nil
		}
		user.mouseMove(p0, p1, p2)
	case EventTypeMouseUp:
		user.mouseUp()
	case EventTypeMouseDown:
		user.mouseDown()
	case EventTypeChangeTool:
		p, err := ToInt(params[0])
		if err != nil {
			user.Error("Invalid tool")
			return nil
		}
		user.changeTool(p != 0)
	case EventTypeChangeColor:
		p0, err0 := ToInt(params[0])
		p1, err1 := ToInt(params[1])
		p2, err2 := ToInt(params[2])
		if err0 != nil || err1 != nil || err2 != nil {
			user.Error("Invalid color")
			return nil
		}
		user.changeColor(p0, p1, p2)
	case EventTypeChangeNickname:
		timestamp := Timestamp()
		nickname, err := ToString(params[0])
		if err != nil {
			user.Error("Invalid nickname")
			return nil
		}
		if utf8.RuneCountInString(nickname) <= MAX_NICKNAME_LENGTH {
			user.changeNickname(nickname, timestamp)
			*event = append(*event, timestamp)
		} else {
			user.Error("Nickname too long")
			return nil
		}
	case EventTypeChatMessage:
		timestamp := Timestamp()
		msg, err := ToString(params[0])
		if err != nil {
			user.Error("Invalid chat message")
			return nil
		}
		user.chatMessage(msg, timestamp)
		*event = append(*event, timestamp)
	}
	return event
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
        close(user.sendData)
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
			if err_msg == "EOF" {
				Log.Printf("user %v left\n", user.UserId)
			} else {
				Log.Printf("user %v was kicked for '%v'\n", user.UserId, err_msg)
			}
			if location != nil {
				user.Location.Quit <- user
			}
			return
		}
	}
}
