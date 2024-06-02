package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	CHAT_DIR   = "data/chat"
	IMAGES_DIR = "data/img"
	LOG_DIR    = "data/log"
	STATIC_DIR = "static"
)

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

var (
	http_port            *int    = flag.Int("p", 80, "Port to listen for http.")
	https_port           *int    = flag.Int("sp", 443, "Port to listen for https.")
	host                 *string = flag.String("host", getenv("DOMAIN", "localhost"), "Website host.")
	certfile             *string = flag.String("cert", "cert", "Certificate file.")
	keyfile              *string = flag.String("key", "key", "Key file.")
	foreground           *bool   = flag.Bool("f", false, "Log on stdout.")
	index_template               = template.Must(template.ParseFiles("templates/index.html"))
	Log                  *log.Logger
	host_with_http_port  string
	host_with_https_port string
)

type JoinRequest struct {
	user       *User
	resultChan chan bool
}

func socket_handler(w http.ResponseWriter, r *http.Request) {

	if redirect_to_host(w, r) {
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		Log.Println(err)
		return
	}

	user := NewUser()
	Log.Printf("New user %v (%v) - (%v)\n", user.UserId, ws.RemoteAddr(), r.UserAgent())
	close_msg := ""
	defer func() {
		ws.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, close_msg),
			time.Now().Add(WRITE_WAIT),
		)
		ws.Close()
		Log.Printf("User %v left (%v)\n", user.UserId, close_msg)
	}()

	// Retrieve the site the user wants to draw over:
	location_url, err := url.QueryUnescape(r.RequestURI[6:]) // skip "/ws?u="
	if err != nil {
		close_msg = fmt.Sprintf("Invalid query: %v", err)
		return
	}

	location := GetLocation(location_url)
	request := &JoinRequest{user, make(chan bool)}
	location.Join <- request
	user_joined := <-request.resultChan

	if !user_joined {
		close_msg = fmt.Sprintf(
			"Too much users at %v, try adding #something at the end of the URL.",
			location_url,
		)
		return
	}

	go user.Sender(ws)
	go user.Receiver(location, ws)
	close_msg = <-user.kick
	location.Quit <- user
}

func signal_handler(c chan os.Signal) {
	Log.Printf("signal %v\n", <-c)
	CloseAllLocations()
	WriteIntToFile(<-userIdGenerator, LAST_USER_ID_FILENAME)
	os.Exit(0)
}

func init() {
	os.MkdirAll(CHAT_DIR, 0777)
	os.MkdirAll(IMAGES_DIR, 0777)
	os.MkdirAll(LOG_DIR, 0777)
	flag.Parse()
	host_with_http_port = fmt.Sprintf("%s:%d", *host, *http_port)
	host_with_https_port = fmt.Sprintf("%s:%d", *host, *https_port)
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

func redirect_to_host(w http.ResponseWriter, r *http.Request) bool {
	if r.Host != *host && r.Host != host_with_http_port && r.Host != host_with_https_port {
		if *http_port == 443 {
			r.URL.Host = *host
		} else {
			r.URL.Host = host_with_https_port
		}
		r.URL.Scheme = "https"
		http.Redirect(w, r, r.URL.String(), 301)
		return true
	}
	return false
}

func redirectHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !redirect_to_host(w, r) {
			h.ServeHTTP(w, r)
		}
	})
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
	img_path := fmt.Sprintf("/%s/", IMAGES_DIR)
	http.HandleFunc("/ws", socket_handler)
	http.Handle("/static/", redirectHandler(http.StripPrefix("/static/", http.FileServer(http.Dir(STATIC_DIR)))))
	http.Handle(img_path, redirectHandler(maxAgeHandler(0, http.StripPrefix(img_path, http.FileServer(http.Dir(IMAGES_DIR))))))
	http.Handle("/", redirectHandler(http.HandlerFunc(index_handler)))

	SignalChan := make(chan os.Signal)
	go signal_handler(SignalChan)
	signal.Notify(SignalChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		Log.Printf("Listening on port %d\n", *https_port)
		err := http.ListenAndServeTLS(fmt.Sprintf(":%d", *https_port), *certfile, *keyfile, nil)
		if err != nil {
			panic("ListenAndServeTLS: " + err.Error())
		}
	}()

	Log.Printf("Listening on port %d\n", *http_port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *http_port), nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
