
window.onload = function() {

    mobileConsole.show(); // https://github.com/B1naryStudio/js-mobile-console
    mobileConsole.options({ showOnError: true, proxyConsole: false, isCollapsed: true, catchErrors: true });
    mobileConsole.toggleCollapsed();

    var cart = document.getElementById("cart");
    var stash = document.getElementById("stash");

    document.getElementById("btogglestash").onclick = function() { 
        stash.style.display = (stash.style.display === "none") ? "block" : "none";
    }

    var ws = new WebSocket(((this.location.protocol === "https:") ? "wss://" : "ws://") + location.host + "/ws");
    ws.onopen = function(evt) {
        console.log("wsOPEN");
        var msg = { command: "getthings" }
        ws.send(JSON.stringify(msg));
    }
    ws.onclose = function(evt) {
        console.log("wsCLOSE");
        ws = null;
    }
    ws.onmessage = function(evt) {
        console.log("wsRESPONSE: " + evt.data);
        var j = JSON.parse(evt.data);
        if (j.command == "update") {
            replaceChildren("cart", j.cart);
            replaceChildren("stash", j.stash);
            // initialize drag drop and listen for sort events
            var s = sortable('.grid', { acceptFrom: ".grid", items: ':not(.header)' })
            for (var i = 0; i<2; i++) s[i].addEventListener('sortupdate', function(e) { sendItems(); });
            document.getElementById("bcartadd").onclick = function() { addNew("cart") }
            document.getElementById("bstashadd").onclick = function() { addNew("stash") }
        }
    }
    ws.onerror = function(evt) {
        console.log("wsERROR: " + evt.data);
    }

    function editThing(n) {
        var res = prompt("New name:", n.innerHTML);
        if (res != null) {
            if (res == "") n.remove()
            else n.innerHTML = res
            sendItems()
        }
    }

    function newThing(name) {
        var e = document.createElement("div")
        e.className = "thing"
        e.innerHTML = name
        e.onclick = function(event) {
            if (event.detail === 1) {
              timer = setTimeout(() => {
                // single click: move to other stack
                var targetid = (event.target.parentNode.id == "cart") ? "stash" : "cart";
                document.getElementById(targetid).appendChild(event.target);
                sendItems()
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
        // e.innerHTML = e.children[0].innerHTML; // remove all but first (header)
        while (e.childNodes.length > 3) e.removeChild(e.lastChild);

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

    function sendItems() {
        var ac = htmlColl2Arr("cart")
        var as = htmlColl2Arr("stash")
        var msg = { command: "updateFromClient", cart: ac, stash: as };
        ws.send(JSON.stringify(msg));
    }

    function addNew(parentid) {
        var n = document.getElementById(parentid).appendChild(newThing("new"));
        editThing(n);
    }

    stash.style.display = "none"; // hide stash on startup

}