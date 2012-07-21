package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"github.com/ugorji/go-msgpack"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"html/template"
	"time"
	"log"
)

var port *int = flag.Int("p", 8000, "Port to listen.")
var sockets map[int]*websocket.Conn
var sockets_lock sync.Mutex
var save_wait sync.WaitGroup
var current_ranking Ranking
var index_template = template.Must(template.ParseFiles("templates/index.html"))
var Log *log.Logger

func sendRecvServer(ws *websocket.Conn) {
	//	fmt.Printf("new connection from %v asking for %v\n", ws.Request().RemoteAddr, ws.Request().RequestURI)
	save_wait.Add(1)
	user := NewUser(ws)
	if user == nil {
		save_wait.Done()
		return
	}
	sockets_lock.Lock()
	sockets[user.UserId] = ws
	sockets_lock.Unlock()
	Log.Println("New user", user.UserId, "joins", user.Location.Url)
	for {
		var buf []byte
		err := websocket.Message.Receive(ws, &buf)
		if err != nil {
			Log.Printf("error while reading socket: %v\n", err)
			break
		}
		//		fmt.Printf("recv:%q\n", buf)
		var v []interface{}
		err = msgpack.Unmarshal([]byte(buf), &v, nil)
		if err != nil {
			Log.Printf("this is not msgpack: '%v'\n", buf)
		} else {
			//			fmt.Printf("Received msgpack encoded: '%v'\n", v)
			user.GotMessage(v)
		}
	}
	user.OnClose()
	ws.Close()
	sockets_lock.Lock()
	delete(sockets, user.UserId)
	sockets_lock.Unlock()
	save_wait.Done()
}

func SignalHandler(c chan os.Signal) {
	Log.Printf("signal %v\n", <-c)
	sockets_lock.Lock()
	for user_id, socket := range sockets {
		Log.Printf("closing connection for user %v\n", user_id)
		socket.Close()
	}
	sockets_lock.Unlock()
	save_wait.Wait()
	// Why do we become a daemon here ?
	Log.Printf("exit\n")
	os.Exit(0)
}

func init() {
	sockets = make(map[int]*websocket.Conn)
	now := time.Now()
	log_file, err := os.Create(now.Format("log/2006-01-02_15:04:05"))
	if err != nil {
		fmt.Println(err)
		panic("Couldn't open log file.")
	}
	Log = log.New(log_file, "", log.LstdFlags)
}

type Website struct {
	Url string
	UserCount int
}

type Ranking []Website

func (r Ranking) Len() int {
    return len(r)
}

func (r Ranking) Less(i, j int) bool {
    return r[i].UserCount > r[j].UserCount  // we went it in the reverse order
}

func (r Ranking) Swap(i, j int) {
    r[i], r[j] = r[j], r[i]
}

func UpdateRanking() {
	var ranking Ranking
	Log.Printf("UpdateRanking wants locations lock.")
	LocationsMutex.Lock()
	Log.Printf("UpdateRanking got locations lock.")
	for _, location := range Locations {
		ranking = append(ranking, Website{Url: location.Url, UserCount: len(location.Users)})
	}
	LocationsMutex.Unlock()
	Log.Printf("UpdateRanking released locations lock.")
	sort.Sort(ranking)
	current_ranking = ranking[:MinInt(len(ranking), 10)]
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	err := index_template.Execute(w, current_ranking)
	if err != nil {
		Log.Printf("Couldn't execute template: %v\n", err)
	}
}

func main() {
	flag.Parse()
	SignalChan := make(chan os.Signal)
	go SignalHandler(SignalChan)
	signal.Notify(SignalChan, os.Interrupt, os.Kill)

	go func() {
		tick := time.Tick(10 * time.Second)
		for _ = range tick {
			UpdateRanking()
		}
	}()

	go func() {
		tick := time.Tick(1 * time.Minute)
		for _ = range tick {
			SaveAllLocations()
		}
	}()

	http.Handle("/ws", websocket.Handler(sendRecvServer))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/", http.HandlerFunc(IndexHandler))
	Log.Printf("Listening on http://localhost:%d/\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		panic("ListenANdServe: " + err.Error())
	}
}
