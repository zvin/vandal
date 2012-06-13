package main

import (
	"code.google.com/p/go.net/websocket"
//    "vandal"
	"flag"
	"fmt"
	"github.com/ugorji/go-msgpack"
	"io"
	"net/http"
	"bytes"
	"os"
	"os/signal"
	"log"
	"sync"
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
//	fmt.Println("sendRecvServer finished")
}

type fileCache struct {
	buf bytes.Buffer
}

func (f *fileCache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write(f.buf.Bytes())
}

func NewCache(fileName string) *fileCache {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("couldn't open file: %v", err)
	}
	ret := &fileCache{}
	_, err = io.Copy(&ret.buf, f)
	if err != nil {
		log.Fatalf("couldn't read file: %v", err)
	}
	return ret
}

func SignalHandler (c chan os.Signal) {
    fmt.Printf("signal %v\n", <-c)
    sockets_lock.Lock()
    for user_id, socket := range sockets {
        fmt.Printf("closing connection for user %v\n", user_id)
        socket.Close()
    }
//    save_wait.Add(len(sockets))
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

	static_files := []string{
	    "index.html", "script.js", "buddy.png", "chat.png", "close.png",
	    "crosshair.png", "eraser.png", "handle.png", "pen.png", "arrow.gif",
	    "hs.png", "cross.gif",
	}
	for _, fname := range static_files {
	    cache := NewCache("static/" + fname)
	    http.Handle("/" + fname, cache)
	    if fname == "index.html" {
	        http.Handle("/", cache)
	    }
	}
	
	http.Handle("/ws", websocket.Handler(sendRecvServer))
	fmt.Printf("http://localhost:%d/\n", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		panic("ListenANdServe: " + err.Error())
	}
}
