package main

import (
	_ "time"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/gorilla/sessions"
	"io"
	"net/http"
	"encoding/json"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"runtime"
	"errors"
)

const MESSAGE_QUEUE_SIZE = 10

var authKey = []byte("somesecretauth")
var store sessions.Store
var pool *Pool
var clients map[int64]*Client

var db *sql.DB
// var tv syscall.Timeval

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

		fmt.Println("match found for ", c1.id, " and ", c2.id)

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

	fmt.Println(session.Values)
	if userid == nil {
		return 0, errors.New("no cookie set")
	} 
	return userid.(int64), nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	db, _ = sql.Open("mysql", "root:pass@/suitup")
	defer db.Close()

	store = sessions.NewCookieStore(authKey)

	pool = newPool()
	clients = make(map[int64]*Client)

	http.HandleFunc("/", mainHandle)

	http.HandleFunc("/login", login)

	http.HandleFunc("/message/check", checkMessage)
	http.HandleFunc("/message/send", sendMessage)

	http.HandleFunc("/chatroom/join", joinChatRoom)
	http.HandleFunc("/chatroom/leave", leaveChatRoom)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("/home/suitup/hackmit/HackMIT2/"))))
	http.ListenAndServe(":8080", nil)
}

type IdQuery struct {
    Id            int64      `json:"id"`

}

func mainHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hey")
}

func joinChatRoom(w http.ResponseWriter, r *http.Request) {
	uid, err := UIDFromSession(w, r)
	handleError(err)

  	fmt.Println("uid: ", uid)
	retChan := make(chan *Room)
	client := &Client{
		id:      uid,
		in:      nil,
		out:     make(chan string, MESSAGE_QUEUE_SIZE),
		retChan: retChan,
	}
	clients[uid] = client
	pool.in <- client

	fmt.Println("added ", uid, " to queue")
	chatroom := <- retChan

	fmt.Fprint(w, "{\"status\":\"success\",\"crid\":", chatroom.id, "}")
}

func leaveChatRoom(w http.ResponseWriter, r *http.Request) {
	uid, _ := UIDFromSession(w, r)
	fmt.Fprint(w, uid)
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	uid, err := UIDFromSession(w, r)
	handleError(err)

	fmt.Println(uid)

	message := r.FormValue("s")

	// message := r.PostFormValue("message")

	client := clients[uid]

	if client != nil {
		client.out <- message
		fmt.Fprint(w, "{\"status\":\"success\"}-", message)
	} else {
		fmt.Fprint(w, "{\"status\":\"failure\"}-", message)
	}	
}

func checkMessage(w http.ResponseWriter, r *http.Request) {
	uid, err := UIDFromSession(w, r)
	handleError(err)

	client := clients[uid]

	if client != nil {
		fmt.Println("client found")
		select {
		case message, ok := <- clients[uid].in:
			fmt.Println("message pulled from channel")
			if ok {
				fmt.Fprint(w, message)
			} else {
				fmt.Fprint(w, "")
			}
		default:
			fmt.Println("")
			fmt.Fprint(w, "")
		}
	
	} else {
		fmt.Println("client not found")
		fmt.Fprint(w, "")
	}
}



func login(w http.ResponseWriter, r *http.Request) {
	inputToken := r.FormValue("access_token")
	if len(inputToken) != 0 {
		uid := GetMe(inputToken)

		// row := db.QueryRow("SELECT id FROM users")
		row := db.QueryRow("SELECT id FROM users WHERE facebook_id=?", string(uid))
		iq := new(IdQuery)
		err := row.Scan(&iq.Id)

		if err != nil {
			_, err = db.Exec("insert into users (facebook_id, username, email, level, points) values (?, ?, ?, 0, 0)", uid, "", "")
			if err != nil {
				fmt.Fprint(w, "{\"status\":\"failure\"}")
				return
			} else {
				row = db.QueryRow("SELECT id FROM users WHERE facebook_id=?", string(uid))
				err = row.Scan(&iq.Id)
				if err != nil {
					fmt.Fprint(w, "{\"status\":\"failure\"}")
					return
				}
			}

		}

		fmt.Println("session-id: ", iq.Id)

		session, _ := store.Get(r, "session")
		session.Values["userid"] = iq.Id
		session.Save(r, w)

		fmt.Println(session.Values)

		fmt.Fprint(w, "{\"status\":\"success\"}")


	// 	if err == nil {
	// 		fmt.Fprint(w, "{\"status\":\"success\",\"uid\":", iq.Id, "}")
	// 	} else {
	// 		_, err = db.Exec("insert into users (facebook_id, username, email, level, points) values (?, ?, ?, 0, 0)", uid, "", "")
	// 		if err == nil {
	// 			row = db.QueryRow("SELECT id FROM users WHERE facebook_id=?", string(uid))
	// 			err = row.Scan(&iq.Id)
	// 			if err == nil {
	// 				fmt.Fprint(w, "{\"status\":\"success\"},\"uid\":", iq.Id, "}")
	// 			} else {
	// 				fmt.Fprint(w, "{\"status\":\"failure\"}")
	// 			}
	// 		} else {
	// 			fmt.Fprint(w, "{\"status\":\"failure\"}")
	// 		}
	// 	}
	// } else {
	// 	fmt.Fprint(w, "{\"status\":\"failure\"}")
	}
}
	

func readHttpBody(response *http.Response) string {
	bodyBuffer := make([]byte, 1000)
	var str string

	count, err := response.Body.Read(bodyBuffer)

	for ; count > 0; count, err = response.Body.Read(bodyBuffer) {

		if err != nil {

		}

		str += string(bodyBuffer[:count])
	}

	return str

}

func getUncachedResponse(uri string) (*http.Response, error) {
	request, err := http.NewRequest("GET", uri, nil)

	if err == nil {
		request.Header.Add("Cache-Control", "no-cache")

		client := new(http.Client)

		return client.Do(request)
	}

	if (err != nil) {
	}
	return nil, err

}

func GetMe(token string) string {
	response, err := getUncachedResponse("https://graph.facebook.com/me?access_token="+token)

	if err == nil {

		var jsonBlob interface{}

		responseBody := readHttpBody(response)

		if responseBody != "" {
			err = json.Unmarshal([]byte(responseBody), &jsonBlob)

			if err == nil {
				jsonObj := jsonBlob.(map[string]interface{})
				return jsonObj["id"].(string)
			}
		}
		return err.Error()
	}

	return err.Error()
}

func handleError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
