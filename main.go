package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"

	auth "github.com/abbot/go-http-auth"
	"github.com/gorilla/websocket"
	"github.com/valyala/fasttemplate"
)

var clients = make(map[*websocket.Conn]string) // connected clients, string is username
var staticmodtime int64 = 0

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Config store gloabl settings
type Config struct {
	Port        int    `json:"port"`
	SSLCertPath string `json:"sslcertpath"`
	SSLKeyPath  string `json:"sslkeypath"`
}

// Message for socket commun
type Message struct {
	Command string   `json:"command"`
	Cart    []string `json:"cart"`
	Stash   []string `json:"stash"`
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
	msg.Command = ""        // message is not stored
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
	switch r.Request.URL.RequestURI() {
	case "/":
		template := MustAsset("index.html")
		t := fasttemplate.New(string(template), "{{", "}}")
		t.Execute(w, map[string]interface{}{
			"staticmodtime": strconv.FormatInt(staticmodtime, 10),
		})
	case "/serviceWorker.js", "/manifest.json": // PWA disabled
		w.WriteHeader(404)
	default:
		http.FileServer(AssetFile()).ServeHTTP(w, &r.Request)
	}
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
			log.Printf("error handleconn, remove client (%v) : %v", ws.RemoteAddr().String(), err)
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

// reload certificate each time https://gist.github.com/KaiserWerk/3c1a5e16c4b85dac1923ecb4d1cbd1dc
func (conf *Config) getCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {

	fmt.Println("GetCertificate() called!")

	caFiles, err := tls.LoadX509KeyPair(conf.SSLCertPath, conf.SSLKeyPath)
	if err != nil {
		return nil, err
	}

	return &caFiles, nil
}

func main() {

	conf := getConfig()

	// get latest static file modTime to avoid js caching https://stackoverflow.com/a/8392506
	for _, v := range _bindata {
		a, err := v()
		if err != nil {
			log.Fatal(err)
		}
		if staticmodtime < a.info.ModTime().Unix() {
			staticmodtime = a.info.ModTime().Unix()
		}
	}

	// http basic auth
	authenticator := auth.NewBasicAuthenticator("wshoppingcart", auth.HtpasswdFileProvider("wshoppingcart-logins.htpasswd"))
	http.HandleFunc("/ws", authenticator.Wrap(handleWS))
	http.HandleFunc("/", authenticator.Wrap(handleFiles))

	log.Printf("Starting server on port %d...", conf.Port)

	var err error
	if conf.SSLKeyPath != "" {
		server := &http.Server{Addr: fmt.Sprintf(":%d", conf.Port), Handler: nil, TLSConfig: &tls.Config{GetCertificate: conf.getCertificate}}
		err = server.ListenAndServeTLS("", "")
	} else {
		server := &http.Server{Addr: fmt.Sprintf(":%d", conf.Port), Handler: nil}
		err = server.ListenAndServe()
	}
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
