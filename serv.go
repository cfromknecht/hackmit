package main

import (
	"crypto/rand"
	"fmt"
	"github.com/gorilla/sessions"
	"io"
	"net/http"
	"encoding/binary"
)

var authKey = []byte("somesecretauth")
var store sessions.Store
var pool *Pool

type Pool struct {
	clients []*Client
	in      chan *Client
	out     chan *Room
}

type Client struct {
	id		int64
	retChan chan *Room
}

type Room struct {
	id 		int64
	client1	*Client
	client2 *Client
}

func (p *Pool) Pair() {
	for {
		fmt.Println("pairing")
		c1 := <- p.in
		fmt.Println("middlePairing")
		c2 := <- p.in
		fmt.Println("donePairing")

		b := make([]byte, 8)
		n, err := io.ReadFull(rand.Reader, b)
		if err != nil || n != 8 {
			return
		}
		crId, _ := binary.Varint(b)

		room := &Room{crId, c1, c2}
		c1.retChan <- room
		c2.retChan <- room
	}
}

func newPool() *Pool {
	pool := &Pool{
		clients:	make([]*Client, 0),
		in:			make(chan *Client),
		out:		make(chan *Room),
	}

	go pool.Pair()

	return pool
}

func main() {
	store = sessions.NewCookieStore(authKey)

	pool = newPool()

	http.HandleFunc("/chatroom/join", joinChatRoom)
	http.HandleFunc("/chatroom/leave", leaveChatRoom)
	http.ListenAndServe(":8080", nil)
}

func joinChatRoom(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userid := session.Values["userid"]
	
	var uid int64
	var b []byte

	if userid == nil {
		b = make([]byte, 8)
		n, err := io.ReadFull(rand.Reader, b)
		if err != nil || n != 8 {
			fmt.Println(err)
			return
		}
		session.Values["userid"] = b
		session.Save(r, w)
	} else {
		b = []byte(userid.([]uint8))
	}
	uid, _ = binary.Varint(b)

	retChan := make(chan *Room)
	client := &Client {uid, retChan}
	pool.in <- client
	fmt.Println("client sent")

	chatroom := <- retChan
	fmt.Println("chatroom received")

	fmt.Fprint(w, "Joined chatroom ", chatroom.id)
}

func leaveChatRoom(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	fmt.Println(session.Values["userid"])
	fmt.Fprint(w, session.Values["userid"])
}
