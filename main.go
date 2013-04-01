package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"github.com/ugorji/go-msgpack"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

const (
	CHAT_DIR   = "chat"
	IMAGES_DIR = "img"
	LOG_DIR    = "log"
	STATIC_DIR = "static"
)

var (
	port           *int  = flag.Int("p", 8000, "Port to listen.")
	foreground     *bool = flag.Bool("f", false, "Log on stdout.")
	index_template       = template.Must(template.ParseFiles("templates/index.html"))
	Log            *log.Logger
)

func socket_handler(ws *websocket.Conn) {

	user := NewUser(ws)

	// Retrieve the site the user wants to draw over:
	location_url, err := url.QueryUnescape(ws.Request().RequestURI[6:]) // skip "/ws?u="
	if err != nil {
		user.Error("Invalid query")
		return
	}

	location := GetLocation(location_url)
	location.Join <- user

	var buffer []byte
	var event []interface{}
	for {
		err := websocket.Message.Receive(ws, &buffer)
		if err != nil {
			if err.Error() == "EOF" {
				Log.Printf("User %v closed connection.\n", user.UserId)
			} else {
				Log.Printf("error while reading socket for user %v: %v\n", user.UserId, err)
			}
			break
		}
		err = msgpack.Unmarshal(buffer, &event, nil)
		if err != nil {
			Log.Printf("this is not msgpack: '%v'\n", buffer)
			user.Error("Invalid message")
		} else {
			location.Message <- UserAndEvent{user, event}
		}
	}
	location.Quit <- user
	ws.Close()
}

func signal_handler(c chan os.Signal) {
	Log.Printf("signal %v\n", <-c)
	CloseAllLocations()
	// Why do we become a daemon here ?
	Log.Printf("exit\n")
	os.Exit(0)
}

func init() {
	os.MkdirAll(CHAT_DIR, 0777)
	os.MkdirAll(IMAGES_DIR, 0777)
	os.MkdirAll(LOG_DIR, 0777)
	flag.Parse()
	now := time.Now()
	var log_file io.Writer
	var err error
	if *foreground == true {
		log_file = os.Stdout
	} else {
		log_file, err = os.Create(LOG_DIR + "/" + now.Format("2006-01-02_15:04:05"))
		if err != nil {
			fmt.Println(err)
			panic("Couldn't open log file.")
		}
	}
	Log = log.New(log_file, "", log.LstdFlags)
}

func index_handler(w http.ResponseWriter, r *http.Request) {
	CurrentlyUsedSites.Mutex.RLock()
	err := index_template.Execute(w, CurrentlyUsedSites.Sites)
	CurrentlyUsedSites.Mutex.RUnlock()
	if err != nil {
		Log.Printf("Couldn't execute template: %v\n", err)
	}
}

func maxAgeHandler(seconds int, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", seconds))
		h.ServeHTTP(w, r)
	})
}

func main() {
	SignalChan := make(chan os.Signal)
	go signal_handler(SignalChan)
	signal.Notify(SignalChan, os.Interrupt, os.Kill)

	http.Handle("/ws", websocket.Handler(socket_handler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(STATIC_DIR))))
	http.Handle("/img/", maxAgeHandler(0, http.StripPrefix("/img/", http.FileServer(http.Dir(IMAGES_DIR)))))
	http.Handle("/", http.HandlerFunc(index_handler))
	Log.Printf("Listening on port %d\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
