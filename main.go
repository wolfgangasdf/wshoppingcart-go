package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"

	auth "github.com/abbot/go-http-auth"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]string) // connected clients, string is username

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

func handleFiles(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	http.FileServer(AssetFile()).ServeHTTP(w, &r.Request)
}

func handleWS(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	ws, err := upgrader.Upgrade(w, &r.Request, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = r.Username // register client websocket

	for { // read loop
		var msg Message
		if err := ws.ReadJSON(&msg); err != nil {
			log.Printf("error read, remove client (%v) : %v", ws.RemoteAddr().String(), err)
			delete(clients, ws)
			break
		}
		log.Print("received cmd: " + msg.Command)
		if msg.Command == "getthings" { // only for this client
			msg2 := thingsRead(r.Username)
			msg2.Command = "update"
			sendHandleError(ws, r.Username, msg2)
		} else if msg.Command == "updateFromClient" {
			thingsWrite(r.Username, &msg)
			msg.Command = "update"
			for cws, cuser := range clients { // send to others
				if cuser == r.Username && cws != ws {
					sendHandleError(cws, r.Username, msg)
				}
			}
		}
	}
}

func main() {

	conf := getConfig()

	authenticator := auth.NewBasicAuthenticator("wshoppingcart", auth.HtpasswdFileProvider("wshoppingcart-logins.htpasswd"))
	http.HandleFunc("/ws", authenticator.Wrap(handleWS))
	http.HandleFunc("/", authenticator.Wrap(handleFiles))

	log.Printf("Starting server on :%d...", conf.Port)

	var err error
	server := &http.Server{Addr: fmt.Sprintf(":%d", conf.Port), Handler: nil}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
