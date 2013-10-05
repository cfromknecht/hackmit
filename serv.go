package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/gorilla/sessions"
	"io"
	"net/http"
)

var authKey = []byte("somesecretauth")
var store sessions.Store
var pool *Pool
var clients map[int64]*Client

type Pool struct {
	in  chan *Client
	out chan *Room
}

type Client struct {
	id      int64
	in      chan string
	out     chan string
	retChan chan *Room
}

type Room struct {
	id      int64
	client1 *Client
	client2 *Client
}

func (p *Pool) Pair() {
	for {
		c1, c2 := <-p.in, <-p.in

		fmt.Println("match found")

		b := make([]byte, 8)
		n, err := io.ReadFull(rand.Reader, b)
		if err != nil || n != 8 {
			return
		}
		crId, _ := binary.Varint(b)

		room := &Room{crId, c1, c2}

		c1.in, c2.in = c2.out, c1.out

		c1.retChan <- room
		c2.retChan <- room
	}
}

func newPool() *Pool {
	pool := &Pool{
		in:  make(chan *Client),
		out: make(chan *Room),
	}

	go pool.Pair()

	return pool
}

func UIDFromSession(w http.ResponseWriter, r *http.Request) (int64, error) {
	session, _ := store.Get(r, "session")
	userid := session.Values["userid"]

	var uid int64
	var b []byte

	if userid == nil {
		b = make([]byte, 8)
		n, err := io.ReadFull(rand.Reader, b)
		if err != nil || n != 8 {
			return 0, err
		}
		session.Values["userid"] = b
		session.Save(r, w)
	} else {
		b = []byte(userid.([]uint8))
	}
	uid, _ = binary.Varint(b)
	return uid, nil
}

func main() {
	store = sessions.NewCookieStore(authKey)

	pool = newPool()
	clients = make(map[int64]*Client)

	http.HandleFunc("/message/check", checkMessage)
	http.HandleFunc("/message/send", sendMessage)
	http.HandleFunc("/chatroom/join", joinChatRoom)
	http.HandleFunc("/chatroom/leave", leaveChatRoom)
	http.ListenAndServe(":8080", nil)
}

func joinChatRoom(w http.ResponseWriter, r *http.Request) {
	uid, err := UIDFromSession(w, r)
	handleError(err)

  	fmt.Println("uid: ", uid)
	retChan := make(chan *Room)
	client := &Client{
		id:      uid,
		in:      nil,
		out:     make(chan string),
		retChan: retChan,
	}
	clients[uid] = client
	pool.in <- client

	fmt.Println("added ", uid, " to queue")
	chatroom := <-retChan

	fmt.Fprint(w, "{\"status\":\"success\",\"crid\":", chatroom.id, "}")
}

func leaveChatRoom(w http.ResponseWriter, r *http.Request) {
	uid, _ := UIDFromSession(w, r)
	fmt.Fprint(w, uid)
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	uid, err := UIDFromSession(w, r)
	handleError(err)

	message := "some string"

	client := clients[uid]

	client.out <- message

	fmt.Fprint(w, "{\"status\":\"success\"}")
}

func checkMessage(w http.ResponseWriter, r *http.Request) {
	uid, err := UIDFromSession(w, r)
	handleError(err)

	message := <-clients[uid].in

	fmt.Fprint(w, message)
}

func handleError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
