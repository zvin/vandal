package main

import (
	"flag"
	"fmt"
	"github.com/garyburd/go-websocket/websocket"
	"html/template"
	"io"
	"log"
	"net"
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

type JoinRequest struct {
	user       *User
	resultChan chan bool
}

func socket_handler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	ws, err := websocket.Upgrade(w, r.Header, nil, 1024, 1024)
	if err != nil {
		log.Println(err)
		return
	}
	defer ws.Close()

	user := NewUser(ws)

	// Retrieve the site the user wants to draw over:
	location_url, err := url.QueryUnescape(r.RequestURI[6:]) // skip "/ws?u="
	if err != nil {
		log.Printf("Invalid query: %v", err)
		return
	}

	location := GetLocation(location_url)
	request := &JoinRequest{user, make(chan bool)}
	location.Join <- request
	user_joined := <-request.resultChan
	if user_joined {
		user.SocketHandler(location)
	} else {
		go func() { // user.Error will block if no SocketHandler is alive
			user.Error("Too much users at this location, try adding #something at the end of the URL.")
		}()
		user.SocketHandler(nil)
	}
}

func signal_handler(c chan os.Signal, listener net.Listener) {
	Log.Printf("signal %v\n", <-c)
	listener.Close() // stop accepting new connections
	CloseAllLocations()
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
	http.HandleFunc("/ws", socket_handler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(STATIC_DIR))))
	http.Handle("/img/", maxAgeHandler(0, http.StripPrefix("/img/", http.FileServer(http.Dir(IMAGES_DIR)))))
	http.Handle("/", http.HandlerFunc(index_handler))
	Log.Printf("Listening on port %d\n", *port)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic("Listen: " + err.Error())
	}

	SignalChan := make(chan os.Signal)
	go signal_handler(SignalChan, listener)
	signal.Notify(SignalChan, os.Interrupt, os.Kill)

	http.Serve(listener, nil)
	WaitLocations()
}
