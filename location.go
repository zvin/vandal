package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/zvin/gocairo"
	"io/ioutil"
	"math"
	"strings"
	"sync"
)

const (
	WIDTH  = 2000
	HEIGHT = 3000
)

var (
	Locations      map[string]*Location
	LocationsMutex sync.RWMutex
)

//func Base64Encode(data []byte) *bytes.Buffer {
func Base64Encode(data string) string {
	bb := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, bb)
	encoder.Write([]byte(data))
	encoder.Close()
	return bb.String()
}

func SaveAllLocations() {
	LocationsMutex.RLock()
	defer LocationsMutex.RUnlock()
	for _, location := range Locations {
		location.Mutex.Lock()
		location.Save()
		location.Mutex.Unlock()
	}
}

type Location struct {
	Url      string
	Users    []*User
	FileName string
	Surface  *cairo.Surface
	delta    []interface{}
	Mutex    sync.RWMutex
}

func init() {
	Locations = make(map[string]*Location)
}

func GetLocation(url string) *Location {
	LocationsMutex.Lock()
	location, present := Locations[url]
	if present {
		return location
	} else {
		location = NewLocation(url)
	}
	LocationsMutex.Unlock()
	return location
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func NewLocation(url string) *Location {
	loc := new(Location)
	loc.Url = url
	Locations[url] = loc
	b64fname := Base64Encode(url)
	b64fname = b64fname[:MinInt(len(b64fname), 251)]
	loc.FileName = "img/" + strings.Replace(b64fname, "/", "_", -1) + ".png" // filename
	fmt.Println(loc.FileName)
	loc.Surface = cairo.NewSurfaceFromPNG(loc.FileName)
	if loc.Surface.SurfaceStatus() != 0 {
		loc.Surface.Finish()
		loc.Surface.Destroy()
		loc.Surface = cairo.NewSurface(cairo.FormatArgB32, WIDTH, HEIGHT)
	}
	loc.Surface.SetSourceRGB(0, 0, 0)
	return loc
}

func (location *Location) AddUser(user *User) {
	location.Users = append(location.Users, user)
}

func (location *Location) RemoveUser(user *User) {
	location.Users = Remove(location.Users, user)
	if len(location.Users) == 0 {
		location.Save()
		LocationsMutex.Lock()
		delete(Locations, location.Url)
		LocationsMutex.Unlock()
	}
}

func MaxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func Distance(x1, y1, x2, y2 int) float64 {
	return math.Sqrt(float64((x1 - x2) ^ 2 + (y1 - y2) ^ 2))
}

func (location *Location) DrawLine(x1, y1, x2, y2, duration, red, green, blue int, use_pen bool) {
	//    fmt.Printf("draw line duration %d\n", duration)
	if duration <= 0 {
		return
	}
	location.delta = append(location.delta, []interface{}{x1, y1, x2, y2, duration, red, green, blue, use_pen})
	d := Distance(x1, y1, x2, y2)
	speed := MaxFloat(d/float64(duration), 1)
	if !use_pen { // not use_pen: use eraser
		location.Surface.SetOperator(cairo.OperatorDestOut)
	} else {
		location.Surface.SetOperator(cairo.OperatorOver)
		location.Surface.SetSourceRGB(float64(red)/255., float64(green)/255., float64(blue)/255.)
	}
	location.Surface.SetLineWidth(1. / (speed * 3))
	location.Surface.MoveTo(float64(x1), float64(y1))
	location.Surface.LineTo(float64(x2), float64(y2))
	location.Surface.Stroke()
}

func (location *Location) GetB64Image() string {
	data, err := ioutil.ReadFile(location.FileName)
	if err != nil {
		fmt.Println("Error reading file ", err)
	}
	return Base64Encode(string(data))
}

func (location *Location) GetDelta() []interface{} {
	// TODO: lock
	//    fmt.Printf("GetDelta returns %v\n", location.delta)
	return location.delta
}

func (location *Location) Save() {
	fmt.Printf("save %s %d\n", location.Url, len(location.delta))
	if len(location.delta) > 0 {
		location.Surface.WriteToPNG(location.FileName) // Output to PNG
		location.delta = nil
	}
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
