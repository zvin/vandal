package main

import (
	"bytes"
	"encoding/base64"
	"github.com/zvin/gocairo"
	"io/ioutil"
	"math"
	"strings"
)

const (
	WIDTH  = 2000
	HEIGHT = 3000
)

var (
	Locations      map[string]*Location
)

func Base64Encode(data string) string {
	bb := &bytes.Buffer{}
	encoder := base64.NewEncoder(base64.StdEncoding, bb)
	encoder.Write([]byte(data))
	encoder.Close()
	return bb.String()
}

func SaveAllLocations() {
	GlobalLock.Lock()
	defer func() {
		GlobalLock.Unlock()
	}()
	for _, location := range Locations {
		location.Save()
	}
}

type Location struct {
	Url      string
	Users    []*User
	FileName string
	Surface  *cairo.Surface
	delta    []interface{}
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
	}
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
	Log.Printf("filename: %v", loc.FileName)
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
		delete(Locations, location.Url)
		location.Surface.Finish()
		location.Surface.Destroy()
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

func (location *Location) GetImageBytes() []byte {
	data, err := ioutil.ReadFile(location.FileName)
	if err != nil {
		Log.Println("Error reading file ", err)
	}
	return data
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
