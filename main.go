package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"github.com/ugorji/go-msgpack"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

var port *int = flag.Int("p", 8000, "Port to listen.")
var sockets map[int]*websocket.Conn
var sockets_lock sync.Mutex
var save_wait sync.WaitGroup

func sendRecvServer(ws *websocket.Conn) {
//	fmt.Printf("new connection from %v asking for %v\n", ws.Request().RemoteAddr, ws.Request().RequestURI)
	save_wait.Add(1)
	user := NewUser(ws)
	sockets_lock.Lock()
	sockets[user.UserId] = ws
	sockets_lock.Unlock()
	fmt.Println(user.UserId, user.Location.Url)
	for {
		var buf []byte
		err := websocket.Message.Receive(ws, &buf)
		if err != nil {
			fmt.Printf("error while reading socket: %v\n", err)
			break
		}
//		fmt.Printf("recv:%q\n", buf)
		var v []interface{}
		err = msgpack.Unmarshal([]byte(buf), &v, nil)
		if err != nil {
			fmt.Printf("this is not msgpack: '%v'\n", buf)
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

func SignalHandler (c chan os.Signal) {
    fmt.Printf("signal %v\n", <-c)
    sockets_lock.Lock()
    for user_id, socket := range sockets {
        fmt.Printf("closing connection for user %v\n", user_id)
        socket.Close()
    }
    sockets_lock.Unlock()
    save_wait.Wait()
    // Why do we become a daemon here ?
    fmt.Printf("exit\n")
    os.Exit(0)
}

func init () {
    sockets = make(map[int]*websocket.Conn)
}

func main() {
	flag.Parse()
    SignalChan := make(chan os.Signal)
    go SignalHandler(SignalChan)
    signal.Notify(SignalChan, os.Interrupt, os.Kill)

    go func() {
        tick := time.Tick(1 * time.Minute)
        for _ = range tick {
            SaveAllLocations()
        }
    }()

	http.Handle("/ws", websocket.Handler(sendRecvServer))
	http.Handle("/", http.FileServer(http.Dir("static")))
	fmt.Printf("http://localhost:%d/\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		panic("ListenANdServe: " + err.Error())
	}
}
