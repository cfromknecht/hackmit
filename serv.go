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
	"database/sql/driver"
	_ "github.com/go-sql-driver/mysql"
	// _ "github.com/ziutek/mymysql/godrv"
	// "github.com/ziutek/mymysql/mysql"
	// _ "github.com/ziutek/mymysql/native"
	// _ "github.com/ziutek/mymysql/thrsafe"
	"runtime"
)

const MESSAGE_QUEUE_SIZE = 10

var authKey = []byte("somesecretauth")
var store sessions.Store
var pool *Pool
var clients map[int64]*Client

var db driver.Conn
// var db = mysql.New("tcp", "", "localhost:3306", "root", "", "suitup")
// var db, _ = sql.Open("mymysql", fmt.Sprintf("%s:%s:%s*%s/%s/%s", "tcp", "localhost", "3306", "suitup", "root", ""))
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
	runtime.GOMAXPROCS(runtime.NumCPU())

	// fmt.Println(1)
	// err := db.Connect()
	// fmt.Println(2)
	// handleError(err)
	// fmt.Println(3)
	// defer db.Close()
	// fmt.Println(4)

	fmt.Println("connecting...")
	// db, err := sql.Open("mysql", fmt.Sprintf("%s:%s:%s*%s/%s/%s", "tcp", "localhost", "3306", "suitup", "root", ""))
	// handleError(err)
	db, err := sql.Open("mysql", "root:@/suitup")
	handleError(err)
	defer db.Close()
	fmt.Println("connected")

	row := db.QueryRow("select id from users")
	rl := new(RequestsList)
	err = row.Scan(&rl.Id)
	fmt.Println(rl.Id)

	store = sessions.NewCookieStore(authKey)

	pool = newPool()
	clients = make(map[int64]*Client)

	http.HandleFunc("/", mainHandle)

	http.HandleFunc("/login", login)

	http.HandleFunc("/message/check", checkMessage)
	http.HandleFunc("/message/send", sendMessage)

	http.HandleFunc("/chatroom/join", joinChatRoom)
	http.HandleFunc("/chatroom/leave", leaveChatRoom)
	http.ListenAndServe(":8080", nil)
}

type RequestsList struct {
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

	message := "some string"

	client := clients[uid]
	if client != nil {
		client.out <- message
		fmt.Fprint(w, "{\"status\":\"success\"}")
	} else {
		fmt.Fprint(w, "{\"status\":\"failure\"}")
	}	
}

func checkMessage(w http.ResponseWriter, r *http.Request) {
	uid, err := UIDFromSession(w, r)
	handleError(err)

	client := clients[uid]
	if client != nil {
		fmt.Println("waiting")
		message := <- clients[uid].in
		fmt.Println("received")
		fmt.Fprint(w, message)
	} else {
		fmt.Fprint(w, "{\"status\":\"failure\"}")
	}
}



func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("login")
	inputToken := r.FormValue("access_token")
	if len(inputToken) != 0 {
		uid := GetMe(inputToken)

		fmt.Println("querying")
		// row := db.QueryRow("SELECT id FROM users WHERE facebook_id='?'", uid)
		// row, _, err := db.QueryFirst("SELECT id FROM users WHERE facebook_id='%s';", uid)
		// handleError(err)

		fmt.Println(uid)

		if w != nil {
			// rl := new(RequestsList)
			// err := row.Scan(&rl.Id)
			fmt.Fprint(w, "{\"status\":\"success\",\"uid\":", "}")
		} else {
			// regStmt, err := db.Prepare("INSERT INTO users (facebook_id, username, email, level, points) VALUES(?, ?, ?, ?, ?);")
			// handleError(err)
			// regStmt.Run(uid, "", "", 0, 0)
			fmt.Fprint(w, "{\"status\":\"success\"}")
		}
	} else {
		fmt.Fprint(w, "{\"status\":\"failure\"}")
	}
}
	

func readHttpBody(response *http.Response) string {

	fmt.Println("Reading body")

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
	fmt.Println("Uncached response GET")
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
	fmt.Println("Getting me")
	response, err := getUncachedResponse("https://graph.facebook.com/me?access_token="+token)

	if err == nil {

		var jsonBlob interface{}

		responseBody := readHttpBody(response)

		fmt.Println("responseboyd", responseBody)

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
