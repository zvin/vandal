package main

import (
	"github.com/zvin/gocairo"
	"strings"
	"sync"
	"time"
)

const (
	WIDTH  = 2000
	HEIGHT = 3000
)

var (
	Locations      map[string]*Location
	LocationsMutex sync.RWMutex
)

type UserAndEvent struct {
	User  *User
	Event []interface{}
}

type Location struct {
	Join         chan *User
	Quit         chan *User
	Message      chan UserAndEvent
	CloseAll     chan bool
	Url          string
	Users        []*User
	FileName     string
	chatFileName string
	Chat         *MessagesLog
	Surface      *cairo.Surface
	delta        []interface{}
}

func init() {
	Locations = make(map[string]*Location)
}

func GetLocation(url string) *Location {
	location, present := Locations[url]
	if present {
		return location
	} else {
		location = NewLocation(url)
		Locations[url] = location
	}
	return location
}

func NewLocation(url string) *Location {
	loc := new(Location)
	loc.Join = make(chan *User)
	loc.Quit = make(chan *User)
	loc.Message = make(chan UserAndEvent)
	loc.CloseAll = make(chan bool)
	loc.Url = url
	b64fname := Base64Encode(url)
	b64fname = b64fname[:MinInt(len(b64fname), 251)]
	b64fname = strings.Replace(b64fname, "/", "_", -1)
	loc.chatFileName = CHAT_DIR + "/" + b64fname + ".gob"
	loc.Chat = OpenMessagesLog(loc.chatFileName)
	loc.FileName = IMAGES_DIR + "/" + b64fname + ".png" // filename
	Log.Printf("filename: %v", loc.FileName)
	loc.Surface = cairo.NewSurfaceFromPNG(loc.FileName)
	if loc.Surface.SurfaceStatus() != 0 {
		loc.Surface.Finish()
		loc.Surface.Destroy()
		loc.Surface = cairo.NewSurface(cairo.FormatArgB32, WIDTH, HEIGHT)
	}
	loc.Surface.SetSourceRGB(0, 0, 0)
	go loc.main()
	return loc
}

func (location *Location) main() {
	save_tick := time.Tick(1 * time.Minute)
	for {
		select {
		case user := <-location.Join:
			location.AddUser(user)
		case user := <-location.Quit:
			location.RemoveUser(user)
			if len(location.Users) == 0 {
				location.Save()
				location.Destroy()
				return
			}
		case message := <-location.Message:
			event := message.User.GotMessage(message.Event)
			if event != nil {
				location.broadcast(message.User, event)
			}
		case <-save_tick:
			location.Save()
		case <-location.CloseAll:
			location.Save()
			location.Destroy()
			return
		}
	}
}

func (location *Location) Destroy() {
	location.Surface.Finish()
	location.Surface.Destroy()
	LocationsMutex.Lock()
	delete(Locations, location.Url)
	LocationsMutex.Unlock()
}

func (location *Location) broadcast(user *User, event []interface{}) {
	// event.insert(1, user.UserId) ...
	event = append(event[:1], append([]interface{}{user.UserId}, event[1:]...)...)
	for _, other := range location.Users {
		other.SendEvent(event)
	}
}

func (location *Location) AddUser(user *User) {
	// Send the list of present users to this user:
	timestamp := Timestamp()
	for _, other := range location.Users {
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
		location.FileName,
		location.GetDelta(),
		location.Chat.GetMessages(),
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
	location.broadcast(user, event) // user is not yet in location.Users, so it will not receive this event.
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
	location.Users = append(location.Users, user)
	location.Chat.AddMessage(timestamp, "", "user "+user.Nickname+" joined")
}

func (location *Location) RemoveUser(user *User) {
	location.Users = Remove(location.Users, user)
	timestamp := Timestamp()
	location.broadcast(user, []interface{}{EventTypeLeave, timestamp})
	location.Chat.AddMessage(timestamp, "", "user "+user.Nickname+" left")
}

func (location *Location) DrawLine(x1, y1, x2, y2, duration, red, green, blue int, use_pen bool) {
	//    fmt.Printf("draw line duration %d\n", duration)
	if duration <= 0 {
		return
	}
	d := Distance(x1, y1, x2, y2)
	if d <= 0 {
		return
	}
	location.delta = append(location.delta, []interface{}{x1, y1, x2, y2, duration, red, green, blue, use_pen})
	speed := d / float64(duration)
	if !use_pen { // not use_pen: use eraser
		location.Surface.SetOperator(cairo.OperatorDestOut)
	} else {
		location.Surface.SetOperator(cairo.OperatorOver)
		location.Surface.SetSourceRGB(float64(red)/255., float64(green)/255., float64(blue)/255.)
	}
	location.Surface.SetLineWidth(1. / (1.3 + (3. * speed)))
	location.Surface.MoveTo(float64(x1), float64(y1))
	location.Surface.LineTo(float64(x2), float64(y2))
	location.Surface.Stroke()
}

func (location *Location) GetDelta() []interface{} {
	return location.delta
}

func (location *Location) Save() {
	if len(location.delta) > 0 {
		Log.Printf("save %s (delta %d)\n", location.Url, len(location.delta))
		location.Surface.WriteToPNG(location.FileName) // Output to PNG
		location.delta = nil
	}
	location.Chat.Save()
}

func Remove(list []*User, value *User) []*User {
	var i int
	var elem *User
	for i, elem = range list {
		if elem == value {
			break
		}
	}
	return append(list[:i], list[i+1:]...)
}
