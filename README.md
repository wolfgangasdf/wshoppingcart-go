# wshoppingcart-go

A very simple multi-user shopping cart app with a "cart" (things to buy) and a "stash" (things in stock), drag'n'drop, realtime synchronized.
It is a test with golang http, ssl with letsencrypt certificate, secure websockets, go-bindata, html5 drag'n'drop,... 

* Click item to move between cart and stash, or drag and drop.
* Double-click to rename.
* Delete item: rename with empty name.
* If the connection is interrupted, you can continue to use it, it will ask what to do if again online. Note that it is not a full progressive web app, you can't start-up offline.

### Run

Add user to htpasswd password file (`-c` creates a new file):

```
htpasswd -c wshoppingcart.htpasswd <username>
htpasswd wshoppingcart.htpasswd <anotheruser>
```

Download the executable, put it on some server that is online 24/7, and run it.

### Settings file 
`wshoppingcart-settings.json`: Leave out the ssl settings to use http, port defaults to 8000:
```
{
"port" : 8000,
"sslcertpath" : "/etc/letsencrypt/live/hostname/fullchain.pem",
"sslkeypath" : "/etc/letsencrypt/live/hostname/privkey.pem"
}
```

The user shopping carts are saved as `wshoppingcart-user-<username>.json`

### build
```
go get -u github.com/go-bindata/go-bindata/v3/... 
# one of:
go-bindata -fs -prefix "static/" static/        # put static files into bindata.go
go-bindata -debug -fs -prefix "static/" static/ # development: use normal files via bindata.go
# one of:
go build
GOOS=linux GOARCH=amd64 go build -o wshoppingcart-linux-amd64 # cross-compile, e.g. for linux
```


# uses

* [html5sortable](http://lukasoppermann.github.io/html5sortable/index.html)
* [gorilla websocket](github.com/gorilla/websocket)
* [js-mobile-console](http://b1narystudio.github.io/js-mobile-console/)
