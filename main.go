package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	auth "github.com/abbot/go-http-auth"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]string) // connected clients, string is username
var broadcast = make(chan Message)             // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Config store gloabl settings
type Config struct {
	Port int `json:"port"`
}

// Message for socket commun
type Message struct {
	Command string   `json:"command"`
	Cart    []string `json:"cart"`
	Stash   []string `json:"stash"`
}

func getConfig() *Config {
	conf := &Config{Port: 8000}
	b, err := ioutil.ReadFile("wshoppingcart-settings.json")
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
	b, err := ioutil.ReadFile(thingsFileName(user))
	if err != nil {
		log.Println("Can't open file: " + err.Error())
	} else {
		if err = json.Unmarshal(b, &msg); err != nil {
			log.Fatal("Can't parse file: " + err.Error())
		}
		msg.Command = ""
	}
	return msg
}

func thingsWrite(user string, msg *Message) {
	msg.Command = ""
	b, err := json.Marshal(msg)
	if err != nil {
		log.Fatal("Can't marshal: " + err.Error())
	} else {
		if err := ioutil.WriteFile(thingsFileName(user), b, 0644); err != nil {
			log.Println("Can't write to file: " + err.Error())
		}
	}

}

func handleStatic(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	http.FileServer(AssetFile()).ServeHTTP(w, &r.Request)
}

func main() {

	conf := getConfig()

	// static
	authenticator := auth.NewBasicAuthenticator("wshoppingcart", auth.HtpasswdFileProvider("wshoppingcart-logins.htpasswd"))
	http.HandleFunc("/", authenticator.Wrap(handleStatic))

	// websocket
	http.HandleFunc("/ws", authenticator.Wrap(handleWSauth))

	log.Printf("Server started on port %d", conf.Port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
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

func handleWSauth(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	handleWS(w, &r.Request, r.Username)
}

func handleWS(w http.ResponseWriter, r *http.Request, user string) {
	log.Printf("handleconnections! ")
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = user

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error handleconn (%v) : %v", ws.RemoteAddr().String(), err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		log.Print("received cmd: " + msg.Command)
		if msg.Command == "getthings" { // only for this client
			msg2 := thingsRead(user)
			msg2.Command = "update"
			sendHandleError(ws, user, msg2)
		} else if msg.Command == "updateFromClient" {
			thingsWrite(user, &msg)
			msg.Command = "update"
			// send to others
			for cws, cuser := range clients {
				if cuser == user && cws != ws {
					sendHandleError(cws, user, msg)
				}
			}
		}
	}
}
