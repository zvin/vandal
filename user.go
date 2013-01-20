package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/ugorji/go-msgpack"
	"net/url"
	"reflect"
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
	UserCount += 1
	user := new(User)
	user.Socket = ws
	user.UserId = UserCount
	location_url, err := url.QueryUnescape(ws.Request().RequestURI[6:])
	if err != nil {
		panic(err)
	}
	user.Location = GetLocation(location_url)
	if len(user.Location.Users) >= MAX_USERS_PER_LOCATION {
		user.Error("Too much users at this location, try adding #something at the end of the URL.")
		return nil
	}
	user.Location.AddUser(user)
	user.UsePen = true
	user.OnOpen()
	return user
}

func EncodeEvent(event []interface{}) []byte {
	result, err := msgpack.Marshal(event)
	if err != nil {
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

func (user *User) SendImage(bytes []byte) {
	if len(bytes) == 0 {
		return
	}
	err := websocket.Message.Send(user.Socket, bytes)
	if err != nil {
		Log.Printf("Couldn't send image to %d: %v\n", user.UserId, err)
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

func (user *User) ChangeNickname(nickname string) {
	user.Nickname = nickname
}

func ToInt(n interface{}) (result int) {
	switch n.(type) {
	case int, int8, int16, int32, int64:
		result = int(reflect.ValueOf(n).Int())
	case uint, uint8, uint16, uint32, uint64:
		result = int(reflect.ValueOf(n).Uint())
	default:
		Log.Printf("not an int: %#v ", n)
		panic("not an int!")
	}
	return result
}

func (user *User) GotMessage(event []interface{}) {
	event_type := ToInt(event[0])
	params := event[1:]
	switch event_type {
	case EventTypeMouseMove:
		user.MouseMove(ToInt(params[0]), ToInt(params[1]), ToInt(params[2]))
	case EventTypeMouseUp:
		user.MouseUp()
	case EventTypeMouseDown:
		user.MouseDown()
	case EventTypeChangeTool:
		user.ChangeTool(params[0].(int8) != 0)
	case EventTypeChangeColor:
		user.ChangeColor(ToInt(params[0]), ToInt(params[1]), ToInt(params[2]))
	case EventTypeChangeNickname:
		user.ChangeNickname(params[0].(string))
	}
	user.Broadcast(event, true)
}

func (user *User) OnOpen() {
	// Send the list of present users to this user:
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
		})
	}
	// Send the image without packing it in msgpack to the new user:
	user.SendImage(user.Location.GetImageBytes())
	// Send the delta between the image and now to the new user:
	user.SendEvent([]interface{}{
		EventTypeWelcome,
		user.Location.GetDelta(),
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
	}
	user.SendEvent(event)
}

func (user *User) OnClose() {
	user.Broadcast([]interface{}{EventTypeLeave}, true)
	user.Location.RemoveUser(user)
}
