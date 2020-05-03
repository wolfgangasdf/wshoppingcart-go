const cacheName = "wshoppingcart-v1";
const assets = [
  ".",
  "/static/app.js",
  "/static/html5sortable.min.js",
  "/static/mobile-console.min.css",
  "/static/mobile-console.min.js",
  "/static/style.css",
];

self.addEventListener("install", installEvent => {
  installEvent.waitUntil(
    caches.open(cacheName).then(cache => {
      cache.addAll(assets);
    })
  );
});

// self.addEventListener("fetch", fetchEvent => {
//   fetchEvent.respondWith(
//     caches.match(fetchEvent.request).then(res => {
//       return res || fetch(fetchEvent.request);
//     })
//   );
// });