package main

// TODO: too many sessions, solve somehow!

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"

	auth "github.com/abbot/go-http-auth"
	"github.com/abraithwaite/jeff"
	"github.com/abraithwaite/jeff/memory"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]jeff.Session) // connected clients TODO: somehow redundant <> jeff.sessions...

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

type server struct {
	jeff *jeff.Jeff
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
		delete(clients, ws) // TODO is jeff session closed?
	}
}

func (s *server) handleLogin(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	log.Printf("handleLogin: r.u=%v url=%v\n", r.Username, r.Request.URL)
	if r.Username != "" {
		err := s.jeff.Set(r.Context(), w, []byte(r.Username), []byte(r.UserAgent()))
		if err != nil {
			log.Println("jeff set error=", err)
			return // TODO err
		}
	}
	sl, _ := s.jeff.SessionsForKey(r.Context(), []byte(r.Username))
	for _, s := range sl {
		log.Printf("  jeff session for user %v: %v\n", r.Username, string(s.Meta))
	}
	http.Redirect(w, &r.Request, "/", http.StatusFound)
}

func (s *server) handleFiles(w http.ResponseWriter, r *http.Request) {
	log.Printf("handlefiles: url=%v\n", r.URL)
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
			log.Printf("error read, remove client (%v) : %v", ws.RemoteAddr().String(), err)
			delete(clients, ws)
			break
		}
		log.Print("received cmd: " + msg.Command)
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

func main() {

	conf := getConfig()

	// jeff manages session authentication also for websocket since http basic can't re-new ws auth without http reload:
	// https://websockets.readthedocs.io/en/stable/topics/authentication.html
	jeffstorage := memory.New()
	s := &server{
		jeff: jeff.New(jeffstorage, jeff.Redirect(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/login", http.StatusFound)
			}))),
	}

	authenticator := auth.NewBasicAuthenticator("wshoppingcart", auth.HtpasswdFileProvider("wshoppingcart-logins.htpasswd"))
	http.Handle("/ws", s.jeff.WrapFunc(s.handleWS))
	http.HandleFunc("/", s.jeff.WrapFunc(s.handleFiles))
	http.HandleFunc("/login", authenticator.Wrap(s.handleLogin))

	log.Printf("Starting server on :%d...", conf.Port)

	var err error
	server := &http.Server{Addr: fmt.Sprintf(":%d", conf.Port), Handler: nil}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
