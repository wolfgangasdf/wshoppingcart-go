package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"

	"crypto/subtle"

	"github.com/abraithwaite/jeff"
	"github.com/abraithwaite/jeff/memory"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]jeff.Session) // connected clients
var usersdebug []string                              // keep track of users that logged in for debug

type server struct {
	jeff *jeff.Jeff
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Config struct {
	Port int `json:"port"`
}

type Message struct {
	Command string   `json:"command"`
	Cart    []string `json:"cart"`
	Stash   []string `json:"stash"`
	Serial  int      `json:"serial"`
}

func getConfig() *Config {
	conf := &Config{Port: 8000}
	b, err := os.ReadFile("wshoppingcart-settings.json")
	if err != nil {
		log.Println("Can't read config: " + err.Error())
	} else {
		if err = json.Unmarshal(b, conf); err != nil {
			log.Fatal("Can't parse config: " + err.Error())
		}
	}
	return conf
}

func readUserDB() map[string]string {
	x := map[string]string{}
	b, err := os.ReadFile("wshoppingcart-users.json")
	if err != nil {
		log.Fatal("Can't read user database: " + err.Error())
	} else {
		if err := json.Unmarshal(b, &x); err != nil {
			log.Fatal("Error parsing user database: ", err)
		}
	}
	return x
}

func thingsFileName(user string) string { return fmt.Sprintf("wshoppingcart-user-%s.json", user) }

func thingsRead(user string) Message {
	msg := Message{Cart: []string{"thingcart"}, Stash: []string{"thingstash"}}
	b, err := os.ReadFile(thingsFileName(user))
	if err != nil {
		log.Println("Can't open file: " + err.Error())
	} else {
		if err = json.Unmarshal(b, &msg); err != nil {
			log.Fatal("Can't parse file: " + err.Error())
		}
	}
	return msg
}

func thingsWrite(user string, msg *Message) {
	msg.Command = ""        // command is not stored
	sort.Strings(msg.Stash) // sort stash
	b, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("Can't marshal: " + err.Error())
	} else {
		if err := os.WriteFile(thingsFileName(user), b, 0644); err != nil {
			log.Println("Can't write to file: " + err.Error())
		}
	}

}

// Send message over websocket, if it fails, close and remove websocket.
func sendHandleError(ws *websocket.Conn, user string, msg Message) {
	log.Printf("sending user %v (%v) message command %v...", user, ws.RemoteAddr().String(), msg.Command)
	err := ws.WriteJSON(msg)
	if err != nil {
		log.Printf("  sending error, close ws and remove from clients: %v", err)
		ws.Close()
		delete(clients, ws)
	}
}

func (s *server) handleLogin(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	pass := r.FormValue("password")
	log.Printf("handleLogin: u=%v p=%v url=%v\n", name, pass, r.URL)
	if name == "" || pass == "" {
		w.WriteHeader(400)
		return
	}
	userMap := readUserDB()
	if subtle.ConstantTimeCompare([]byte(userMap[name]), []byte(pass)) == 1 {
		log.Println("correct password!")
		err := s.jeff.Set(r.Context(), w, []byte(name), []byte(r.UserAgent()))
		if err != nil {
			log.Println("jeff set error=", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		usersdebug = append(usersdebug, name)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *server) handleLogout(w http.ResponseWriter, r *http.Request) {
	log.Printf("handlelogout: url=%v\n", r.URL)
	for cws, s := range clients { // send to others for same user!
		if bytes.Equal(s.Token, jeff.ActiveSession(r.Context()).Token) {
			log.Printf("  closing ws: %v\n", cws.RemoteAddr())
			cws.Close()
		}
	}
	s.jeff.Clear(r.Context(), w)

	http.Redirect(w, r, "/p/login.html", http.StatusFound)
}

func (s *server) handleFiles(w http.ResponseWriter, r *http.Request) {
	log.Printf("handlefiles: url=%v\n", r.URL.Path)
	http.FileServer(AssetFile()).ServeHTTP(w, r)
}

func (s *server) handleWS(w http.ResponseWriter, r *http.Request) {
	log.Printf("handlews: url=%v\n", r.URL)
	sess := jeff.ActiveSession(r.Context())
	user := string(sess.Key)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws error: %v", err)
		return
	}
	defer ws.Close()

	clients[ws] = sess // register client websocket

	for { // read loop
		var msg Message
		if err := ws.ReadJSON(&msg); err != nil {
			log.Printf("ws: error read, remove client if exists (%v) : %v", ws.RemoteAddr().String(), err)
			delete(clients, ws)
			break
		}
		log.Print("ws: received cmd: " + msg.Command)
		if msg.Command == "getthings" { // only for this client
			msg2 := thingsRead(user)
			msg2.Command = "update"
			sendHandleError(ws, user, msg2)
		} else if msg.Command == "updateFromClient" {
			thingsWrite(user, &msg)
			msg.Command = "update"
			for cws, s := range clients { // send to others for same user!
				if string(s.Key) == user && cws != ws {
					sendHandleError(cws, user, msg)
				}
			}
		}
	}
}

func (s *server) printDebugOnKey() {
	reader := bufio.NewReader(os.Stdin)
	for {
		reader.ReadRune()
		log.Println("debug:")
		for cws, js := range clients { // send to others for same user!
			log.Printf("  debug: ws client: wsremote=%v jeffkey=%v jefftoken=%v\n", cws.RemoteAddr(), js.Key, js.Token)
		}

		for _, name := range usersdebug {
			sl, _ := s.jeff.SessionsForKey(nil, []byte(name))
			for _, s := range sl {
				log.Printf("  debug: have jeff session for user %v: %v\n", name, s.Token)
			}
		}
	}
}

func main() {

	conf := getConfig()

	var jeffstorage = memory.New()

	s := &server{
		jeff: jeff.New(jeffstorage, jeff.Redirect(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/p/login.html", http.StatusFound)
			}))),
	}

	http.HandleFunc("/p/", s.handleFiles)
	http.HandleFunc("/logout", s.jeff.WrapFunc(s.handleLogout))
	http.HandleFunc("/login", s.handleLogin)
	http.Handle("/ws", s.jeff.WrapFunc(s.handleWS))
	http.Handle("/", s.jeff.WrapFunc(s.handleFiles)) // jeff redirects to login if needed

	log.Println("Press any key to show debug information...")
	go s.printDebugOnKey()

	log.Printf("Starting server on :%d...", conf.Port)
	server := &http.Server{Addr: fmt.Sprintf(":%d", conf.Port), Handler: nil}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
