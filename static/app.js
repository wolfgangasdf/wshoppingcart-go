
window.onload = function() {

    const LSLASTSYNCMS = "lastsyncms";
    const LSSEND2SERVER = "send2server";
    const SYNCTIMEOUTS = 60; // after this offline time in seconds, ask before pushing local changes.

    mobileConsole.show(); // https://github.com/B1naryStudio/js-mobile-console
    mobileConsole.options({ showOnError: true, proxyConsole: false, isCollapsed: true, catchErrors: true });
    mobileConsole.toggleCollapsed();

    var cart = document.getElementById("cart");
    var stash = document.getElementById("stash");
    var help = document.getElementById("help");
    var wait = document.getElementById("wait");

    stash.style.display = "none";
    help.style.display = "none";
    cart.addEventListener('sortupdate', function(e) { sendItems(); });
    stash.addEventListener('sortupdate', function(e) { sendItems(); });

    document.getElementById("btogglestash").onclick = function() { 
        stash.style.display = (stash.style.display === "none") ? "block" : "none";
    }

    document.getElementById("btogglehelp").onclick = function() { 
        help.style.display = (help.style.display === "none") ? "block" : "none";
    }

    document.getElementById("bcartadd").onclick = function() { addNew("cart") }
    document.getElementById("bstashadd").onclick = function() { addNew("stash") }

    function updateLastsync() { window.localStorage.setItem(LSLASTSYNCMS, Date.now()); }

    var wsWasOpen = false;
    var ws = null
    function startWebsocket() {
        ws = new WebSocket(((this.location.protocol === "https:") ? "wss://" : "ws://") + location.host + "/ws");
        ws.onopen = function(evt) {
            console.log("wsOPEN ");
            var ls = window.localStorage.getItem(LSSEND2SERVER);
            var lslastsyncms = parseInt(window.localStorage.getItem(LSLASTSYNCMS))
            if (ls) {
                var secs = Math.round((Date.now() - lslastsyncms) / 1000);
                console.log(`  secs = ${secs}`);
                if (secs < SYNCTIMEOUTS || confirm(`Last sync with server is ${secs}s ago, should I push to server (cancel: poll)?`)) {
                    ws.send(ls);
                }
            }
            window.localStorage.removeItem(LSSEND2SERVER);
            wait.style.display = "none"
            ws.send(JSON.stringify({ command: "getthings" }));
            wsWasOpen = true;
        }
        ws.onclose = function(evt) { // is also called after unsuccessful connection attempt!
            console.log("wsCLOSE ", ws.readyState, wsWasOpen);
            if (wsWasOpen) updateLastsync();
            wait.style.display = "block"
            setTimeout(function(){startWebsocket()}, 500);
            ws = null;
            wsWasOpen = false;
        }
        ws.onmessage = function(evt) {
            console.log("wsRESPONSE!");
            var j = JSON.parse(evt.data);
            if (j.command == "update") {
                replaceChildren("cart", j.cart);
                replaceChildren("stash", j.stash);
                // (re)initialize drag drop and listen for sort events
                var s = sortable('.grid', { acceptFrom: ".grid", items: ':not(.header)' })
                updateLastsync();
            }
        }
        ws.onerror = function(evt) {
            console.log("wsERROR: " + evt.data);
        }
    }
    startWebsocket()

    function send2server(m) {
        if (ws == null || ws.readyState !== WebSocket.OPEN) {
            window.localStorage.setItem(LSSEND2SERVER, m);
            console.log("can't send, queue! ws=", ws);
        } else {
            ws.send(m);
            updateLastsync();
        }
    }

    function sendItems() {
        var ac = htmlColl2Arr("cart")
        var as = htmlColl2Arr("stash")
        var msg = { command: "updateFromClient", cart: ac, stash: as };
        send2server(JSON.stringify(msg))
    }

    function editThing(n) {
        var res = prompt("New name:", n.innerHTML);
        if (res == null || res == "") {
            n.remove()
        } else {
            n.innerHTML = res
        }
        sendItems()
    }

    function newThing(name) {
        var e = document.createElement("div")
        e.className = "thing"
        e.innerHTML = name
        e.onclick = function(event) {
            if (event.detail === 1) {
              timer = setTimeout(() => {
                // single click: move to other stack
                if (event.target.parentNode != null) {
                    var targetid = (event.target.parentNode.id == "cart") ? "stash" : "cart";
                    document.getElementById(targetid).appendChild(event.target);
                    sendItems()
                }
              }, 200)
            }
          }
        e.ondblclick = function(event) {
            clearTimeout(timer)
            // double click: rename
            var n = event.target
            editThing(n);
        }
        return e
    }

    function replaceChildren(id, arr) {
        var e = document.getElementById(id);
        while (e.childNodes.length > 3) e.removeChild(e.lastChild); // remove all but header
        for (var i = 0; i < arr.length; i++) {
            e.appendChild(newThing(arr[i]))
        }
    }

    function htmlColl2Arr(id) {
        cis = document.getElementById(id).children;
        var arr = [];
        for (i = 0; i < cis.length; i++) {
            if (cis[i].classList.contains("thing")) arr.push(cis[i].innerHTML);
        }
        return arr
    }

    function addNew(parentid) {
        var n = document.getElementById(parentid).appendChild(newThing(""));
        editThing(n);
    }

    // PWA stuff
    // if ('serviceWorker' in navigator){
    //     navigator.serviceWorker.register('/serviceWorker.js').then(function(registration){
    //         console.log('service worker registration succeeded:',registration);
    //     },
    //     function(error){
    //         console.log('service worker registration failed:',error);
    //     });
    // } else {
    //     console.log('service workers are not supported.');
    // }
}