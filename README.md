# wshoppingcart-go

A multi-user shopping cart app with a "cart" and a "stash", drag'n'drop, realtime synchronized.
It is a test with golang http, ssl with letsencrypt certificate, secure websockets, go-bindata, html5 drag'n'drop,... 

* Rename item: if name is empty it is removed.

### Add user to htpasswd password file (`-c` creates a new file):

```
htpasswd -c wshoppingcart.htpasswd <username>
htpasswd wshoppingcart.htpasswd <anotheruser>
```

### Settings file
Leave out the ssl settings to use http, port defaults to 8000:
```
{
"port" : 8000,
"sslcertpath" : "/etc/letsencrypt/live/quphotonics.com/fullchain.pem",
"sslkeypath" : "/etc/letsencrypt/live/quphotonics.com/privkey.pem"
}
```

### run
```
go build && ./wshoppingcart-go
```


### package static files
```
go get -u github.com/go-bindata/go-bindata/... # but make sure to get 3.1.3, grrr
go-bindata -fs -prefix "static/" static/
go-bindata -debug -fs -prefix "static/" static/ # use normal files
```

### build & run
```
go build && ./wshoppingcart-go
```

### build for linux
GOOS=linux GOARCH=amd64 go build -o wshoppingcart-linux-amd64


### test websocket security in js console chrome (incognito mode)
```
var ws = new WebSocket("ws://localhost:8000/ws");
var ws = new WebSocket("wss://quphotonics.org:8000/ws");
```
this must fail!

# uses

* [html5sortable](http://lukasoppermann.github.io/html5sortable/index.html)
* [gorilla websocket](github.com/gorilla/websocket)
