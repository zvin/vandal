package main

import (
	"fmt"
	"github.com/zvin/gocairo"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	WIDTH  = 2000
	HEIGHT = 3000
)

var (
	locations struct {
		sync.RWMutex
		m map[string]*Location
	}
	locationsWait      sync.WaitGroup
	CurrentlyUsedSites LockableWebsiteSlice
)

type UserAndEvent struct {
	User  *User
	Event *[]interface{}
}

type Location struct {
	Join         chan *JoinRequest
	Quit         chan *User
	Message      chan *UserAndEvent
	Close        chan bool
	UserCount    chan int
	Url          string
	Chat         *MessagesLog
	users        []*User
	fileName     string
	chatFileName string
	surface      *cairo.Surface
	delta        []interface{}
}

func init() {
	locations.m = make(map[string]*Location)
	go func() {
		tick := time.Tick(10 * time.Second)
		for _ = range tick {
			update_currently_used_sites()
		}
	}()
}

func GetLocation(url string) *Location {
	locations.Lock()
	defer locations.Unlock()
	location, present := locations.m[url]
	if present {
		// a location already exists
		count := <-location.UserCount
		if count == 0 {
			delete(locations.m, location.Url) // this location has closed, we need to recreate it
		} else {
			return location
		}
	}
	location = newLocation(url)
	locations.m[url] = location
	return location
}

func CloseAllLocations() {
	locations.Lock()
	for _, loc := range locations.m {
		loc.Close <- true
		delete(locations.m, loc.Url)
	}
	locations.Unlock()
}

func WaitLocations() {
	locationsWait.Wait()
}

func update_currently_used_sites() {
	var sites []*Website
	locations.Lock()
	for _, location := range locations.m {
		count := <-location.UserCount
		if count > 0 {
			site := new(Website)
			site.Url = location.Url
			site.UserCount = count
			sites = append(sites, site)
		} else {
			delete(locations.m, location.Url) // no more users, close location
		}
	}
	locations.Unlock()
	SortWebsites(sites)
	sites = sites[:MinInt(len(sites), 10)]
	for _, site := range sites {
		site.Label = TruncateString(TryQueryUnescape(site.Url), 30)
		site.UserCountLabel = fmt.Sprintf(
			"%d %s",
			site.UserCount,
			Pluralize("user", site.UserCount),
		)
	}
	CurrentlyUsedSites.Mutex.Lock()
	CurrentlyUsedSites.Sites = sites
	CurrentlyUsedSites.Mutex.Unlock()
}

func newLocation(url string) *Location {
	loc := new(Location)
	loc.Join = make(chan *JoinRequest)
	loc.Quit = make(chan *User)
	loc.Message = make(chan *UserAndEvent)
	loc.Close = make(chan bool)
	loc.UserCount = make(chan int)
	loc.Url = url
	b64fname := Base64Encode(url)
	b64fname = b64fname[:MinInt(len(b64fname), 251)]
	b64fname = strings.Replace(b64fname, "/", "_", -1)
	loc.chatFileName = CHAT_DIR + "/" + b64fname + ".gob"
	loc.Chat = OpenMessagesLog(loc.chatFileName)
	loc.fileName = IMAGES_DIR + "/" + b64fname + ".png" // filename
	Log.Printf("filename: %v", loc.fileName)
	loc.surface = cairo.NewSurfaceFromPNG(loc.fileName)
	if loc.surface.SurfaceStatus() != 0 {
		loc.surface.Finish()
		loc.surface.Destroy()
		loc.surface = cairo.NewSurface(cairo.FormatArgB32, WIDTH, HEIGHT)
	}
	loc.surface.SetSourceRGB(0, 0, 0)
	go loc.main()
	return loc
}

func (location *Location) main() {
	locationsWait.Add(1)
	defer locationsWait.Done()
	save_tick := time.Tick(1 * time.Minute)
	for {
		select {
		case request := <-location.Join:
			if len(location.users) >= MAX_USERS_PER_LOCATION {
				request.resultChan <- false
			} else {
				Log.Println("New user", request.user.UserId, "joins", TryQueryUnescape(location.Url))
				request.resultChan <- true
				location.addUser(request.user)
			}
		case user := <-location.Quit:
			location.removeUser(user)
			if len(location.users) == 0 {
				location.save()
				location.destroy()
				// this location will be removed from the locations map by GetLocation or update_currently_used_sites
			}
		case message := <-location.Message:
			event := location.UserGotEvent(message.User, message.Event)
			if event != nil {
				location.broadcast(message.User, event)
			}
		case <-save_tick:
			location.save()
		case <-location.Close:
			// Called by CloseAllLocations when we need to quit
			location.save()
			location.destroy()
			return
		case location.UserCount <- len(location.users):
			if len(location.users) == 0 {
				// We have 0 users and will be deleted from locations map now:
				return // stop processing events for this location
			}
		}
	}
}

func (location *Location) destroy() {
	location.surface.Finish()
	location.surface.Destroy()
}

func (location *Location) broadcast(user *User, event *[]interface{}) {
	// event.insert(1, user.UserId) ...
	*event = append((*event)[:1], append([]interface{}{user.UserId}, (*event)[1:]...)...)
	data, err := encodeEvent(event)
	if err != nil {
		return
	}
	for _, other := range location.users {
		other.SendData(data)
	}
}

func (location *Location) addUser(user *User) {
	// Send the list of present users to this user:
	timestamp := Timestamp()
	for _, other := range location.users {
		user.SendEvent(&[]interface{}{
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
	user.SendEvent(&[]interface{}{
		EventTypeWelcome,
		location.fileName,
		location.getDelta(),
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
	location.broadcast(user, &event) // user is not yet in location.users, so it will not receive this event.
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
	user.SendEvent(&event)
	location.users = append(location.users, user)
	location.Chat.AddMessage(timestamp, "", "user "+user.Nickname+" joined")
}

func (location *Location) removeUser(user *User) {
	location.users = remove(location.users, user)
	timestamp := Timestamp()
	location.broadcast(user, &[]interface{}{EventTypeLeave, timestamp})
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
		location.surface.SetOperator(cairo.OperatorDestOut)
	} else {
		location.surface.SetOperator(cairo.OperatorOver)
		location.surface.SetSourceRGB(float64(red)/255., float64(green)/255., float64(blue)/255.)
	}
	location.surface.SetLineWidth(1. / (1.3 + (3. * speed)))
	location.surface.MoveTo(float64(x1), float64(y1))
	location.surface.LineTo(float64(x2), float64(y2))
	location.surface.Stroke()
}

func (location *Location) getDelta() []interface{} {
	return location.delta
}

func (location *Location) save() {
	if len(location.delta) > 0 {
		Log.Printf("save %s (delta %d)\n", location.Url, len(location.delta))
		location.surface.WriteToPNG(location.fileName) // Output to PNG
		location.delta = nil
	}
	location.Chat.Save()
}

func (location *Location) UserGotEvent(user *User, event *[]interface{}) *[]interface{} {
	event_type, err := ToInt((*event)[0])
	if err != nil {
		user.Error("Invalid event type")
		return nil
	}
	params := (*event)[1:]
	switch event_type {
	case EventTypeMouseMove:
		x, err0 := ToInt(params[0])
		y, err1 := ToInt(params[1])
		duration, err2 := ToInt(params[2])
		if err0 != nil || err1 != nil || err2 != nil {
			user.Error("Invalid mouse move")
			return nil
		}
		if user.MouseIsDown {
			location.DrawLine(
				user.PositionX, user.PositionY, // origin
				x, y, // destination
				duration,                                       // duration
				user.ColorRed, user.ColorGreen, user.ColorBlue, // color
				user.UsePen, // pen or eraser
			)
		}
		user.mouseMove(x, y, duration)
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
		nickname, err := ToString(params[0])
		if err != nil {
			user.Error("Invalid nickname")
			return nil
		}
		if utf8.RuneCountInString(nickname) > MAX_NICKNAME_LENGTH {
			user.Error("Nickname too long")
			return nil
		}
		timestamp := Timestamp()
		location.Chat.AddMessage(timestamp, "", user.Nickname+" is now known as "+nickname)
		user.changeNickname(nickname)
		*event = append(*event, timestamp)
	case EventTypeChatMessage:
		msg, err := ToString(params[0])
		if err != nil {
			user.Error("Invalid chat message")
			return nil
		}
		timestamp := Timestamp()
		location.Chat.AddMessage(timestamp, user.Nickname, msg)
		*event = append(*event, timestamp)
	}
	return event
}

func remove(list []*User, value *User) []*User {
	var i int
	var elem *User
	for i, elem = range list {
		if elem == value {
			break
		}
	}
	return append(list[:i], list[i+1:]...)
}
