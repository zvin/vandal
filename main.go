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
	"os"
	"os/signal"
	"sync"
	"time"
)

const (
	CHAT_DIR   = "chat"
	IMAGES_DIR = "img"
	LOG_DIR    = "log"
	STATIC_DIR = "static"
)

var (
	GlobalLock     sync.Mutex
	port           *int  = flag.Int("p", 8000, "Port to listen.")
	foreground     *bool = flag.Bool("f", false, "Log on stdout.")
	save_wait      sync.WaitGroup
	index_template = template.Must(template.ParseFiles("templates/index.html"))
	Log            *log.Logger
)

func socket_handler(ws *websocket.Conn) {
	save_wait.Add(1)
	user := NewUser(ws)
	if user == nil {
		save_wait.Done()
		return
	}
	Log.Println("New user", user.UserId, "joins", user.Location.Url)
	for {
		var buf []byte
		err := websocket.Message.Receive(ws, &buf)
		if err != nil {
			Log.Printf("error while reading socket for user %v: %v\n", user.UserId, err)
			break
		}
		var v []interface{}
		err = msgpack.Unmarshal(buf, &v, nil)
		if err != nil {
			Log.Printf("this is not msgpack: '%v'\n", buf)
		} else {
			Log.Println("GotMessage", "want Lock")
			GlobalLock.Lock()
			Log.Println("GotMessage", "got Lock")
			user.GotMessage(v)
			GlobalLock.Unlock()
			Log.Println("GotMessage", "released Lock")
		}
	}
	Log.Println("OnClose", "want Lock")
	GlobalLock.Lock()
	Log.Println("OnClose", "got Lock")
	user.OnClose()
	GlobalLock.Unlock()
	Log.Println("OnClose", "released Lock")
	ws.Close()
	save_wait.Done()
}

func signal_handler(c chan os.Signal) {
	Log.Printf("signal %v\n", <-c)
	GlobalLock.Lock()
	for _, loc := range Locations {
		for _, user := range loc.Users {
			user.Socket.Close()
		}
	}
	GlobalLock.Unlock()
	save_wait.Wait()
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
	err := index_template.Execute(w, CurrentlyUsedSites)
	if err != nil {
		Log.Printf("Couldn't execute template: %v\n", err)
	}
}

func main() {
	SignalChan := make(chan os.Signal)
	go signal_handler(SignalChan)
	signal.Notify(SignalChan, os.Interrupt, os.Kill)

	go func() {
		tick := time.Tick(10 * time.Second)
		for _ = range tick {
			UpdateCurrentlyUsedSites()
		}
	}()

	go func() {
		tick := time.Tick(1 * time.Minute)
		for _ = range tick {
			SaveAllLocations()
		}
	}()

	http.Handle("/ws", websocket.Handler(socket_handler))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(STATIC_DIR))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir(IMAGES_DIR))))
	http.Handle("/", http.HandlerFunc(index_handler))
	Log.Printf("Listening on http://localhost:%d/\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		panic("ListenANdServe: " + err.Error())
	}
}
