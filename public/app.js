
window.onload = function() {

    mobileConsole.show(); // https://github.com/B1naryStudio/js-mobile-console
    mobileConsole.options({ showOnError: true, proxyConsole: false, isCollapsed: true, catchErrors: true });
    mobileConsole.toggleCollapsed();

    var ws = new WebSocket("ws://" + location.host + "/ws");
    ws.onopen = function(evt) {
        console.log("wsOPEN");
        var msg = { command: "getthings" }
        ws.send(JSON.stringify(msg));
    }
    ws.onclose = function(evt) {
        console.log("wsCLOSE");
        ws = null;
    }

    function replaceChildren(id, arr) {
        var e = document.getElementById(id);
        e.innerHTML = ""
        for (var i = 0; i < arr.length; i++) {
            var newe = document.createElement("div")
            newe.innerHTML = arr[i]
            // TODO doesn't work properly...
            newe.onclick = function(a) { // move to other (don't use newe here...)
                console.log("a: ", a.target.innerHTML, " parent: ", a.target.parentNode.id);
                var targetid = (a.target.parentNode.id == "cart") ? "stash" : "cart";
                document.getElementById(targetid).appendChild(newe);
                sendItems()
            }
            e.appendChild(newe)
        }
    }

    ws.onmessage = function(evt) {
        console.log("wsRESPONSE: " + evt.data);
        var j = JSON.parse(evt.data);
        if (j.command == "update") {
            replaceChildren("cart", j.cart);
            replaceChildren("stash", j.stash);
            sortable(".grid", { acceptFrom: ".grid" });
        }
    }

    ws.onerror = function(evt) {
        console.log("wsERROR: " + evt.data);
    }

    function htmlColl2Arr(id) {
        cis = document.getElementById(id).children;
        var arr = [];
        for (i = 0; i < cis.length; i++) arr.push(cis[i].innerHTML);
        return arr
    }

    function sendItems() {
        var ac = htmlColl2Arr("cart")
        var as = htmlColl2Arr("stash")
        var msg = { command: "updateFromClient", cart: ac, stash: as };
        ws.send(JSON.stringify(msg));
    }

    // listen for sort events
    for (var i = 0; i < 2; i++) sortable('.grid')[i].addEventListener('sortupdate', function(e) {
        sendItems();
    });

    function addNewNotify(id) {
        var newe = document.createElement("div")
        newe.innerHTML = "new"
        document.getElementById(id).appendChild(newe)
        sendItems()
    }
    this.document.getElementById("bcartadd").onclick = function() {
        addNewNotify("cart")
    }
    this.document.getElementById("bstashadd").onclick = function() {
        addNewNotify("stash")
    }

}