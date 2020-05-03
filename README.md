# wshoppingcart-go

A multi-user shopping cart app with a "cart" and a "stash", drag'n'drop, realtime synchronized.
It is a test with golang http, ssl with letsencrypt certificate, secure websockets, go-bindata, html5 drag'n'drop,... 

* Rename item: if name is empty it is removed.
* If the connection is interrupted, you can continue to use it, it will ask what to do if again online.

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

### build
```
go get -u github.com/go-bindata/go-bindata/v3/... 
go-bindata -fs -prefix "static/" static/        # put static files into bindata.go
go-bindata -debug -fs -prefix "static/" static/ # development: use normal files via bindata.go
go build && ./wshoppingcart-go
```

### cross-compile, e.g. for linux
GOOS=linux GOARCH=amd64 go build -o wshoppingcart-linux-amd64


# uses

* [html5sortable](http://lukasoppermann.github.io/html5sortable/index.html)
* [gorilla websocket](github.com/gorilla/websocket)
* [js-mobile-console](http://b1narystudio.github.io/js-mobile-console/)
