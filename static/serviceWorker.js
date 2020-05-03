const cacheName = "wshoppingcart-v1";
const assets = [
    "/",
    "/app.js",
    "/html5sortable.min.js",
    "/mobile-console.min.css",
    "/mobile-console.min.js",
    "/style.css",
];

self.addEventListener("install", installEvent => {
    installEvent.waitUntil(
        caches.open(cacheName).then(cache => {
            cache.addAll(assets);
        })
    );
});

self.addEventListener("fetch", fetchEvent => {
    fetchEvent.respondWith(
        caches.match(fetchEvent.request).then(res => {
            return res || fetch(fetchEvent.request);
        })
    );
});