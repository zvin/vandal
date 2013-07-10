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
	locations          = make(map[string]*Location)
	locationsWait      sync.WaitGroup
	locationsMutex     sync.RWMutex
	CurrentlyUsedSites LockableWebsiteSlice
)

type UserAndEvent struct {
	User  *User
	Event []interface{}
}

type Location struct {
	Join         chan *JoinRequest
	Quit         chan *User
	Message      chan UserAndEvent
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
	go func() {
		tick := time.Tick(10 * time.Second)
		for _ = range tick {
			update_currently_used_sites()
		}
	}()
}

func GetLocation(url string) *Location {
	locationsMutex.Lock()
	defer locationsMutex.Unlock()
	location, present := locations[url]
	if present {
		// a location already exists
		count := <-location.UserCount
		if count == 0 {
			delete(locations, location.Url) // this location has closed, we need to recreate it
		} else {
			return location
		}
	}
	location = newLocation(url)
	locations[url] = location
	return location
}

func CloseAllLocations() {
	locationsMutex.Lock()
	for _, loc := range locations {
		loc.Close <- true
		delete(locations, loc.Url)
	}
	locationsMutex.Unlock()
}

func WaitLocations() {
	locationsWait.Wait()
}

func update_currently_used_sites() {
	var sites []Website
	locationsMutex.Lock()
	for _, location := range locations {
		count := <-location.UserCount
		if count > 0 {
			sites = append(sites, Website{Url: location.Url, UserCount: count})
		} else {
			delete(locations, location.Url) // no more users, close location
		}
	}
	locationsMutex.Unlock()
	SortWebsites(sites)
	CurrentlyUsedSites.Mutex.Lock()
	CurrentlyUsedSites.Sites = sites[:MinInt(len(sites), 10)]
	CurrentlyUsedSites.Mutex.Unlock()
}

func newLocation(url string) *Location {
	loc := new(Location)
	loc.Join = make(chan *JoinRequest)
	loc.Quit = make(chan *User)
	loc.Message = make(chan UserAndEvent)
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
				Log.Println("New user", request.user.UserId, "joins", location.Url)
				location.addUser(request.user)
				request.resultChan <- true
			}
		case user := <-location.Quit:
			location.removeUser(user)
			if len(location.users) == 0 {
				location.save()
				location.destroy()
				// this location will be removed from the locations map by GetLocation or update_currently_used_sites
				return // stop processing events for this location
			}
		case message := <-location.Message:
			event := message.User.GotMessage(message.Event)
			if event != nil {
				location.broadcast(message.User, event)
			}
		case <-save_tick:
			location.save()
		case <-location.Close:
			for _, user := range location.users {
				user.Socket.Close()
			}
		case location.UserCount <- len(location.users):
		}
	}
}

func (location *Location) destroy() {
	location.surface.Finish()
	location.surface.Destroy()
}

func (location *Location) broadcast(user *User, event []interface{}) {
	// event.insert(1, user.UserId) ...
	event = append(event[:1], append([]interface{}{user.UserId}, event[1:]...)...)
	for _, other := range location.users {
		other.SendEvent(event)
	}
}

func (location *Location) addUser(user *User) {
	user.Location = location
	// Send the list of present users to this user:
	timestamp := Timestamp()
	for _, other := range location.users {
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
	location.broadcast(user, event) // user is not yet in location.users, so it will not receive this event.
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
	location.users = append(location.users, user)
	location.Chat.AddMessage(timestamp, "", "user "+user.Nickname+" joined")
}

func (location *Location) removeUser(user *User) {
	location.users = remove(location.users, user)
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
