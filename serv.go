package main

import (
	_ "time"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"io"
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"runtime"
	"errors"
	"html/template"
	"os/exec"
)

var templates = template.Must(template.ParseFiles("index.html"))

const MESSAGE_QUEUE_SIZE = 10

const STATUS_FAILURE = "{\"status\":\"failure\"}"

var authKey = []byte("somesecretauth")
var store sessions.Store

var pool *Pool
var clients map[int64]*Client

var db *sql.DB

type Pool struct {
	in  chan *Client
	out chan *Room
}

type Client struct {
	id      int64
	otherid int64
	in      chan string
	out     chan string
	retChan chan *Room
}

type Room struct {
	id      []byte
	client1 *Client
	client2 *Client
	Question *Question
}

type Question struct {
	Id 	int64			`json:"id"`
	Title string		`json:"title"`
	Body string			`json:"body"`
	Difficulty int 		`json:"diff"`
}



func (p *Pool) Pair() {
	for {
		c1, c2 := <-p.in, <-p.in

		for c1.id == c2.id {
			c2 = <- p.in
		}

		fmt.Println("match found for ", c1.id, " and ", c2.id)

		b := make([]byte, 32)
		n, err := io.ReadFull(rand.Reader, b)
		if err != nil || n != 32 {
			return
		}

		row := db.QueryRow("SELECT * FROM questions ORDER BY RAND()")
		q := new(Question)
		err = row.Scan(&q.Id, &q.Title, &q.Body, &q.Difficulty)
		if err !=  nil {
			fmt.Println(err)
		}

		room := &Room{b, c1, c2, q}

		c1.otherid, c2.otherid = c2.id, c1.id
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

	http.HandleFunc("/question/new", newQuestion)
	http.HandleFunc("/question/submit", testCode)

	http.HandleFunc("/chatroom/join", joinChatRoom)
	http.HandleFunc("/chatroom/leave", leaveChatRoom)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("/home/suitup/hackmit/assets/"))))
	http.ListenAndServe(":8080", nil)
}

func mainHandle(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}

func joinChatRoom(w http.ResponseWriter, r *http.Request) {
	uid, err := UIDFromSession(w, r)
	handleError(err)

	fmt.Println("join ", uid)

	retChan := make(chan *Room)
	client := &Client{
		id:      uid,
		in:      nil,
		out:     make(chan string, MESSAGE_QUEUE_SIZE),
		retChan: retChan,
	}

	clients[uid] = client
	pool.in <- client

	chatroom := <- retChan

	fmt.Println("joinChatRoom-chatroom.id: ", chatroom.id)
	
	qjs, err := json.Marshal(chatroom.Question)

	fmt.Println("qjs: ", string(qjs))

	fmt.Fprint(w, "{\"status\":\"success\",\"crid\":\"", asciify(chatroom.id), "\",\"question\":", string(qjs), "}")
}

func asciify(ba []byte) string {
	ret := make([]byte, len(ba))
	for i, b := range ba {
		ret[i] = (b % 26) + 97
	}
	return string(ret)
}

func testCode(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	code := r.FormValue("submission")
	_ = r.FormValue("cvid")
	qid := "1"
	app := "secure.sh"
	// out1, err1 := exec.Command("/usr/bin/secure", qid, code).Output()
	// if err1 != nil {
 //    	fmt.Println(err1)
	// }
	// fmt.Printf("The date is %s\n", out1)
	out, err := exec.Command("bash", "-c", app + " " + qid + " '" + code + "'").Output()
    if err != nil {
    	fmt.Println(err)
    	return
    }
    fmt.Println(string(out))
    fmt.Fprint(w, "{\"status\":\"success\", \"data\": [", string(out), "]}" )
    return
}

func leaveChatRoom(w http.ResponseWriter, r *http.Request) {
	uid, _ := UIDFromSession(w, r)
	client := clients[uid]

	if client != nil {
		fmt.Println("leaving")
		delete(clients, client.otherid)
		delete(clients, uid)
	}

	fmt.Println("leave ", uid)

	fmt.Fprint(w, "{\"status\":\"success\"}")
}

func sendMessage(w http.ResponseWriter, r *http.Request) {
	uid, err := UIDFromSession(w, r)
	handleError(err)

	fmt.Println("send ", uid)

	r.ParseForm()
	message := r.FormValue("s")

	client := clients[uid]

	if client != nil {
		client.out <- message
		fmt.Fprint(w, "{\"status\":\"success\"}")
	} else {
		fmt.Fprint(w, STATUS_FAILURE)
	}	
}

func checkMessage(w http.ResponseWriter, r *http.Request) {
	uid, err := UIDFromSession(w, r)
	handleError(err)

	fmt.Println("check ", uid)

	client := clients[uid]

	if client != nil {
		select {
		case message, ok := <- client.in:
			if ok {
				fmt.Fprint(w, "{\"status\":\"success\",\"s\":\"", message, "\"}")
			} else {
				delete(clients, uid)
				fmt.Fprint(w, "{\"status\":\"failure\"}")
			}
		default:
			fmt.Fprint(w, "{\"status\":\"success\",\"s\":\"\"}")
		}
	
	} else {
		fmt.Fprint(w, "{\"status\":\"failure\",\"s\":\"\"}")
	}
}

type User struct {
    Id            	int64      	`json:"id"`
    FacebookId  	string		`json:fbid"`
    Username 		string		`json:username"`
    Email 			string		`json:email"`
    Level 			int64		`json:lvl"`
    Score 			int64		`json:score"`
}

func newQuestion(w http.ResponseWriter, r *http.Request) {
	row := db.QueryRow("SELECT * FROM questions ORDER BY RAND()")
	q := new(Question)
	err := row.Scan(&q.Id, &q.Title, &q.Body, &q.Difficulty)

	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w, STATUS_FAILURE)
	}

	b, err := json.Marshal(q)
	if err != nil {
		fmt.Println(err)
		fmt.Fprint(w, STATUS_FAILURE)
	}

	fmt.Fprint(w, string(b))
}


func login(w http.ResponseWriter, r *http.Request) {
	inputToken := r.FormValue("access_token")
	if len(inputToken) != 0 {
		uid := GetMe(inputToken)

		// row := db.QueryRow("SELECT id FROM users")
		row := db.QueryRow("SELECT id FROM users WHERE facebook_id=?", string(uid))
		user := new(User)
		err := row.Scan(&user.Id)

		if err != nil {
			_, err = db.Exec("INSERT INTO users (facebook_id, username, email, level, points) VALUES (?, ?, ?, 0, 0)", uid, "", "")
			if err != nil {
				fmt.Fprint(w, STATUS_FAILURE)
				return
			} else {
				row = db.QueryRow("SELECT id FROM users WHERE facebook_id=?", string(uid))
				err = row.Scan(&user.Id)
				if err != nil {
					fmt.Fprint(w, STATUS_FAILURE)
					return
				}
			}

		}

		session, _ := store.Get(r, "session")
		session.Values["userid"] = user.Id
		session.Save(r, w)

		fmt.Fprint(w, "{\"status\":\"success\"}")
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
