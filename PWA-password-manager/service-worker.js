self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open('my-cache').then((cache) => {
      return cache.addAll([
        '/icons',
        '/index.html',
        '/client.js',
        '/offline.html',
      ]);
    })
  );
  console.log('Service Worker installed');
});

self.addEventListener('fetch', (event) => {
  event.respondWith(
    caches.match(event.request).then((cachedResponse) => {
      if (cachedResponse) {
        console.log('returning cache response');
        return cachedResponse;
      }
      return fetch(event.request).catch(() => {
        console.log('returning offline');
        return caches.match('/offilne.html');
      });
    })
  );
});
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== 'my-cache') {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});
