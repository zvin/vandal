package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/ugorji/go-msgpack"
	"net/url"
	"reflect"
	"errors"
	"strconv"
)

const MAX_USERS_PER_LOCATION = 20

var UserCount int

type User struct {
	Socket      *websocket.Conn
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
}

func NewUser(ws *websocket.Conn) *User {
	Log.Println("NewUser", "want Lock")
	GlobalLock.Lock()
	Log.Println("NewUser", "got Lock")
	defer func() {
		GlobalLock.Unlock()
		Log.Println("NewUser", "released Lock")
	}()
	Log.Println("NewUser", "1")
	UserCount += 1
	user := new(User)
	Log.Println("NewUser", "2")
	user.Socket = ws
	user.UserId = UserCount
	user.Nickname = strconv.Itoa(user.UserId)
	location_url, err := url.QueryUnescape(ws.Request().RequestURI[6:])
	Log.Println("NewUser", "3")
	if err != nil {
		Log.Println("NewUser", "panic")
		panic(err)
	}
	Log.Println("NewUser", "4")
	user.Location = GetLocation(location_url)
	Log.Println("NewUser", "5")
	if len(user.Location.Users) >= MAX_USERS_PER_LOCATION {
		Log.Println("NewUser", "too much")
		user.Error("Too much users at this location, try adding #something at the end of the URL.")
		return nil
	}
	Log.Println("NewUser", "6")
	user.Location.AddUser(user)
	Log.Println("NewUser", "7")
	user.UsePen = true
	Log.Println("NewUser", "8")
	user.OnOpen()
	Log.Println("NewUser", "9")
	return user
}

func EncodeEvent(event []interface{}) []byte {
	result, err := msgpack.Marshal(event)
	if err != nil {
		Log.Printf("Couldn't encode event '%v'\n", event)
		panic(err)
	}
	return result
}

func (user *User) SendEvent(event []interface{}) {
	//    fmt.Printf("sending %v\n", event)
	//    fmt.Printf("sending %v to %d %#v\n", EncodeEvent(event), user.UserId, user.Socket)
	err := websocket.Message.Send(user.Socket, EncodeEvent(event))
	if err != nil {
		Log.Printf("Couldn't send to %d: %v\n", user.UserId, err)
		user.Socket.Close()
	}
}

func (user *User) Error(description string) {
	user.SendEvent([]interface{}{
		EventTypeError,
		description,
	})
	user.Socket.Close()
}

func (user *User) Broadcast(event []interface{}, include_myself bool) {
	//    fmt.Printf("users %v\n", user.Location.Users)
	// event.insert(1, user.UserId) ...
	event = append(event[:1], append([]interface{}{user.UserId}, event[1:]...)...)
	for _, other := range user.Location.Users {
		if other != nil {
			//            fmt.Printf("user %v\n", other)
			if !include_myself && other == user {
				continue
			}
			other.SendEvent(event)
		}
	}
}

func (user *User) MouseMove(x int, y int, duration int) {
	//    fmt.Printf("mouse move\n")
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

func (user *User) MouseUp() {
	user.MouseIsDown = false
}

func (user *User) MouseDown() {
	user.MouseIsDown = true
}

func (user *User) ChangeTool(use_pen bool) {
	user.UsePen = use_pen
}

func (user *User) ChangeColor(red, green, blue int) {
	user.ColorRed = red
	user.ColorGreen = green
	user.ColorBlue = blue
}

func (user *User) ChangeNickname(nickname string, timestamp int64) {
	user.Location.Chat.AddMessage(timestamp, "", user.Nickname + " is now known as " + nickname)
	user.Nickname = nickname
}

func (user *User) ChatMessage(msg string, timestamp int64) {
	user.Location.Chat.AddMessage(timestamp, user.Nickname, msg)
}

func ToInt(n interface{}) (result int, err error) {
	switch n.(type) {
	case int, int8, int16, int32, int64:
		result = int(reflect.ValueOf(n).Int())
	case uint, uint8, uint16, uint32, uint64:
		result = int(reflect.ValueOf(n).Uint())
	default:
		Log.Printf("ToInt, not an int: %#v ", n)
		err = errors.New("Not an int")
	}
	return result, err
}

func (user *User) GotMessage(event []interface{}) {
	event_type, err := ToInt(event[0])
	if err != nil{
		user.Error("Invalid event type")
		return
	}
	params := event[1:]
	switch event_type {
	case EventTypeMouseMove:
		p0, err0 := ToInt(params[0])
		p1, err1 := ToInt(params[1])
		p2, err2 := ToInt(params[2])
		if err0 != nil || err1 != nil || err2 != nil {
			user.Error("Invalid mouse move")
			return
		}
		user.MouseMove(p0, p1, p2)
	case EventTypeMouseUp:
		user.MouseUp()
	case EventTypeMouseDown:
		user.MouseDown()
	case EventTypeChangeTool:
		user.ChangeTool(params[0].(int8) != 0)
	case EventTypeChangeColor:
		p0, err0 := ToInt(params[0])
		p1, err1 := ToInt(params[1])
		p2, err2 := ToInt(params[2])
		if err0 != nil || err1 != nil || err2 != nil {
			user.Error("Invalid color")
			return
		}
		user.ChangeColor(p0, p1, p2)
	case EventTypeChangeNickname:
		timestamp := Timestamp()
		user.ChangeNickname(params[0].(string), timestamp)
		event = append(event, timestamp)
	case EventTypeChatMessage:
		timestamp := Timestamp()
		user.ChatMessage(params[0].(string), timestamp)
		event = append(event, timestamp)
	}
	user.Broadcast(event, true)
}

func (user *User) OnOpen() {
	// Send the list of present users to this user:
	timestamp := Timestamp()
	for _, other := range user.Location.Users {
		user.SendEvent([]interface{}{
			EventTypeJoin,
			other.UserId,
			[]interface{}{other.PositionX, other.PositionY},
			[]interface{}{other.ColorRed, other.ColorGreen, other.ColorBlue},
			other.MouseIsDown,
			false, // not you
			other.Nickname,
			other.UsePen,
			0, // no timestamp
		})
	}
	// Send the delta between the image and now to the new user:
	user.SendEvent([]interface{}{
		EventTypeWelcome,
		user.Location.FileName,
		user.Location.GetDelta(),
		user.Location.Chat.GetMessages(),
	})
	// Send this new user to other users:
	event := []interface{}{
		EventTypeJoin,
		[]interface{}{user.PositionX, user.PositionY},
		[]interface{}{user.ColorRed, user.ColorGreen, user.ColorBlue},
		user.MouseIsDown,
		false, // not you
		user.Nickname,
		user.UsePen,
		timestamp,
	}
	user.Broadcast(event, false)
	// Send this user to himself
	event = []interface{}{
		EventTypeJoin,
		user.UserId,
		[]interface{}{user.PositionX, user.PositionY},
		[]interface{}{user.ColorRed, user.ColorGreen, user.ColorBlue},
		user.MouseIsDown,
		true, // you
		user.Nickname,
		user.UsePen,
		timestamp,
	}
	user.SendEvent(event)
	user.Location.Chat.AddMessage(timestamp, "", "user " + user.Nickname + " joined")
}

func (user *User) OnClose() {
	timestamp := Timestamp()
	user.Broadcast([]interface{}{EventTypeLeave, timestamp}, true)
	user.Location.Chat.AddMessage(timestamp, "", "user " + user.Nickname + " left")
	user.Location.RemoveUser(user)
}
