
"use strict"

window.onload = function() {

    const LSLASTSERVERSERIAL = "lastserverserial"
    const LSHAVEUNSENTCHANGES = "haveunsentchanges"
    const LONGPRESSDELETEMS = 600;
    const TOUCHDRAGDELAYMS = 200;

    var cart = document.getElementById("cart");
    var stash = document.getElementById("stash");
    var help = document.getElementById("help");
    var wait = document.getElementById("wait");

    var ws = null
    var timer = null

    stash.style.display = "none";
    help.style.display = "none";

    mobileConsole.show(); // https://github.com/B1naryStudio/js-mobile-console
    mobileConsole.options({ showOnError: true, proxyConsole: false, isCollapsed: true, catchErrors: true });
    mobileConsole.toggleCollapsed();

    Sortable.create(cart, {
        delay: TOUCHDRAGDELAYMS,
        delayOnTouchOnly: true,
        draggable: ".thing",
        touchStartThreshold: 10,
        onSort: function() {
            sendItems();
        }
    })

    document.getElementById("btogglestash").onclick = function() { 
        stash.style.display = (stash.style.display === "none") ? "block" : "none";
    }
    document.getElementById("btogglehelp").onclick = function() { 
        help.style.display = (help.style.display === "none") ? "block" : "none";
    }
    document.getElementById("bcartadd").onclick = function() { addNew("cart") }
    document.getElementById("bstashadd").onclick = function() { addNew("stash") }

    function startWebsocket() {
        var wsurl = ((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + "/ws"
        console.log("opening ws:", wsurl)
        ws = new WebSocket(wsurl);
        ws.onopen = function() {
            wait.style.display = "none"
            ws.send(JSON.stringify({ command: "getthings" }));
        }
        ws.onclose = function() { // is also called after unsuccessful connection attempt!
            console.log("ws onclose: ", ws.readyState);
            wait.style.display = "block"
            setTimeout(function(){startWebsocket()}, 500);
            ws = null;
        }
        ws.onmessage = function(evt) {
            var j = JSON.parse(evt.data);
            if (j.command == "update") {
                if (window.localStorage.getItem(LSHAVEUNSENTCHANGES) != null) {
                    window.localStorage.removeItem(LSHAVEUNSENTCHANGES);
                    if (parseInt(window.localStorage.getItem(LSLASTSERVERSERIAL)) == j.serial || confirm(`Local and server data modified, push to server (cancel: poll)?`)) { 
                        sendItems()
                        return
                    }
                }
                window.localStorage.setItem(LSLASTSERVERSERIAL, j.serial)
                replaceChildren("cart", j.cart);
                replaceChildren("stash", j.stash);
            }
        }
        ws.onerror = function(evt) {
            console.log("ws onerror: " + evt.data);
        }
    }

    function send2server(m, newserial) {
        if (ws == null || ws.readyState !== WebSocket.OPEN) {
            window.localStorage.setItem(LSHAVEUNSENTCHANGES, "1");
            console.log("send2server: can't send, flag unsaved changes! ws=", ws);
        } else {
            ws.send(m);
            window.localStorage.setItem(LSLASTSERVERSERIAL, newserial) // if send was successful, store new server serial
        }
    }

    function sendItems() {
        var ac = htmlColl2Arr("cart")
        var as = htmlColl2Arr("stash")
        var newserial = parseInt(window.localStorage.getItem(LSLASTSERVERSERIAL)) + 1
        var msg = { command: "updateFromClient", cart: ac, stash: as, serial: newserial };
        send2server(JSON.stringify(msg), newserial)
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

    startWebsocket()

}