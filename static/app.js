
"use strict"

window.onload = function() {

    const LSLASTSYNCMS = "lastsyncms";
    const LSSEND2SERVER = "send2server";
    const SYNCTIMEOUTS = 60; // after this offline time in seconds, ask before pushing local changes.
    const LONGPRESSDELETEMS = 600;
    const TOUCHDRAGDELAYMS = 200;

    mobileConsole.show(); // https://github.com/B1naryStudio/js-mobile-console
    mobileConsole.options({ showOnError: true, proxyConsole: false, isCollapsed: true, catchErrors: true });
    mobileConsole.toggleCollapsed();

    var cart = document.getElementById("cart");
    var stash = document.getElementById("stash");
    var help = document.getElementById("help");
    var wait = document.getElementById("wait");

    var timer = null

    var sortable = Sortable.create(cart, {
        delay: TOUCHDRAGDELAYMS,
        delayOnTouchOnly: true,
        draggable: ".thing",
        touchStartThreshold: 10,
        onSort: function() {
            sendItems();
        }
    })

    stash.style.display = "none";
    help.style.display = "none";

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
        var wsurl = ((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + "/ws"
        console.log("opening ws:", wsurl)
        ws = new WebSocket(wsurl);
        ws.onopen = function() {
            var ls = window.localStorage.getItem(LSSEND2SERVER);
            var lslastsyncms = parseInt(window.localStorage.getItem(LSLASTSYNCMS))
            if (ls) {
                var secs = Math.round((Date.now() - lslastsyncms) / 1000);
                if (secs < SYNCTIMEOUTS || confirm(`Last sync with server is ${secs}s ago, should I push to server (cancel: poll)?`)) {
                    ws.send(ls);
                }
            }
            window.localStorage.removeItem(LSSEND2SERVER);
            wait.style.display = "none"
            ws.send(JSON.stringify({ command: "getthings" }));
            wsWasOpen = true;
        }
        ws.onclose = function() { // is also called after unsuccessful connection attempt!
            console.log("ws onclose: ", ws.readyState, wsWasOpen);
            if (wsWasOpen) updateLastsync();
            wait.style.display = "block"
            setTimeout(function(){startWebsocket()}, 500);
            ws = null;
            wsWasOpen = false;
        }
        ws.onmessage = function(evt) {
            var j = JSON.parse(evt.data);
            if (j.command == "update") {
                replaceChildren("cart", j.cart);
                replaceChildren("stash", j.stash);
                updateLastsync();
            }
        }
        ws.onerror = function(evt) {
            console.log("ws onerror: " + evt.data);
        }
    }
    startWebsocket()

    function send2server(m) {
        if (ws == null || ws.readyState !== WebSocket.OPEN) {
            window.localStorage.setItem(LSSEND2SERVER, m);
            console.log("send2server: can't send, enqueue! ws=", ws);
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
        var res = prompt("Rename (empty to delete):", n.innerHTML);
        if (res == "") {
            n.remove()
        } else if (res != null)  {
            n.innerHTML = res
        }
        sendItems()
    }

    function newThing(name) {
        var e = document.createElement("div")
        e.className = "thing"
        e.innerHTML = name
        var longpresstimeout = null;
        e.onclick = function(event) { // single click: move to other stack
            if (longpresstimeout != null) return; // just to make sure, if order is down,click,up
            if (event.detail === 1) {
              timer = setTimeout(() => {
                if (event.target.parentNode != null) {
                    var targetid = (event.target.parentNode.id == "cart") ? "stash" : "cart";
                    document.getElementById(targetid).appendChild(event.target);
                    sendItems()
                }
              }, 200)
            }
          }
        e.ondblclick = function(event) { // double click: rename item
            clearTimeout(timer)
            var n = event.target
            editThing(n)
        }
        e.addEventListener('touchstart', (event) => {  // mobile long press: delete
            longpresstimeout = setTimeout(function(){
                // should cancel accidental drag here but can't be done: https://github.com/SortableJS/Sortable/issues/264
                if (confirm('Really delete "' + event.target.innerHTML + '"?')) {
                    event.target.remove()
                    sendItems()
                }
            }, LONGPRESSDELETEMS)
        })
        function cancellongpress() {
            clearTimeout(longpresstimeout)
            longpresstimeout = null
        }
        e.addEventListener('touchend', cancellongpress);
        e.addEventListener('touchmove', cancellongpress);
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
        var cis = document.getElementById(id).children;
        var arr = [];
        for (var i = 0; i < cis.length; i++) {
            if (cis[i].classList.contains("thing")) arr.push(cis[i].innerHTML);
        }
        return arr
    }

    function addNew(parentid) {
        var n = document.getElementById(parentid).appendChild(newThing(""));
        editThing(n);
    }

}