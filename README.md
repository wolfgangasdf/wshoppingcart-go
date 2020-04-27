# wshoppingcart-go

TODO
* test apache ssl proxy!!! if this doesn't work...
* go-bindata or so!

```
go build
./whoppingcart-go
```

Create password file (`-c` creates a new file):

```
htpasswd -c wshoppingcart.htpasspwd <username>
```

### package static files
go get -u github.com/go-bindata/go-bindata/...
go-bindata -fs -prefix "static/" static/

### build & run
go build && ./wshoppingcart-go



### test websocket security in js console chrome (incognito mode)
```
var ws = new WebSocket("ws://localhost:8000/ws");
ws.send('{ "command": "update" }');
```
this must fail if no http basic auth, and work with!


Then point your browser to http://localhost:8000

# uses

http://lukasoppermann.github.io/html5sortable/index.html
