# wshoppingcart-go

A multi-user shopping cart app, realtime synchronized with "cart" and "stash".
It was a test with golang http, ssl with letsencrypt certificate, websockets, go-bindata, html5 drag'n'drop... 

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
ws.send('{ "command": "update" }');
```
this must fail if no http basic auth, and work with!

# uses

* [html5sortable](http://lukasoppermann.github.io/html5sortable/index.html)
* [gorilla websocket](github.com/gorilla/websocket)
