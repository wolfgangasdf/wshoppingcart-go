# wshoppingcart-go

A very simple multi-user shopping cart app with a "cart" (things to buy) and a "stash" (things in stock), drag'n'drop, realtime synchronized between clients.
It was just a test with golang http, websockets, go-bindata, html5 drag'n'drop, but now I use it daily.

* Click/tap item to move between cart and stash
* Double-click/tap to rename item (delete if empty)
* Long-touch: delete item
* Cart items can be reordered by drag'n'drop
* If the connection is interrupted, you can continue to use it, if there is a change conflict, it will ask what to do if again online. Note that it is not a full progressive web app, you can't start-up offline

### Run

Run it behind a SSL reverse proxy (websocket is at /ws), see below! 

Add user to htpasswd password file (`-c` creates a new file):

```
htpasswd -c wshoppingcart-logins.htpasswd <username>
htpasswd wshoppingcart-logins.htpasswd <anotheruser>
```

Download the executable, put it on some server that is online 24/7, and run it.


### Settings file 
`wshoppingcart-settings.json`: Leave out the ssl settings to use http, port defaults to 8000:
```
{
"port" : 8000,
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

### Example apache2 reverse proxy configuration with SSL
```
# this is reverse proxy from https://wshoppingcart.example.org to http://127.0.0.1:8000

<IfModule mod_ssl.c>
<VirtualHost *:443>
ServerAdmin example@gmail.com
ErrorLog ${APACHE_LOG_DIR}/error.log
CustomLog ${APACHE_LOG_DIR}/access.log combined

ServerName wshoppingcart.example.org
SSLCertificateFile /etc/letsencrypt/live/example/fullchain.pem
SSLCertificateKeyFile /etc/letsencrypt/live/example/privkey.pem
Include /etc/letsencrypt/options-ssl-apache.conf

# hsts
Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains"

# reverse proxy
SSLProxyEngine on
ProxyPreserveHost On
ProxyPass / http://127.0.0.1:8000/ retry=1
ProxyPassReverse / http://127.0.0.1:8000/
ProxyRequests Off

# websocket https://httpd.apache.org/docs/2.4/mod/mod_proxy_wstunnel.html
ProxyPass /ws ws://127.0.0.1:8000/ws
ProxyPassReverse /ws ws://127.0.0.1:8000/ws
RewriteEngine on
RewriteCond %{HTTP:Upgrade} websocket [NC]
RewriteCond %{HTTP:Connection} upgrade [NC]
RewriteRule /(.*) "ws://localhost:8000/$1" [P,L]

</VirtualHost>
</IfModule>
```

# uses

* [SortableJS](https://github.com/SortableJS/Sortable)
* [gorilla websocket](github.com/gorilla/websocket)
* [js-mobile-console](http://b1narystudio.github.io/js-mobile-console/)
