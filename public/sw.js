// Service Worker for PWA — E-Pesantren
const CACHE_NAME = 'epesantren-v4';
const PRECACHE = ['/icon-192.png', '/icon-512.png', '/style.css'];

self.addEventListener('install', e => {
  e.waitUntil(caches.open(CACHE_NAME).then(c => c.addAll(PRECACHE)));
  self.skipWaiting();
});

self.addEventListener('activate', e => {
  e.waitUntil(caches.keys().then(keys =>
    Promise.all(keys.filter(k => k !== CACHE_NAME).map(k => caches.delete(k)))
  ));
  self.clients.claim();
});

self.addEventListener('fetch', e => {
  const url = new URL(e.request.url);
  // Biarkan browser handle request CDN / cross-origin (mencegah bug cache limit di iOS)
  if (url.origin !== location.origin) return;
  // Jangan cache request API
  if (e.request.url.includes('/api/')) return;
  
  // Network-first, fallback to cache
  e.respondWith(
    fetch(e.request).then(r => {
      if (r.ok) {
        const clone = r.clone();
        caches.open(CACHE_NAME).then(c => c.put(e.request, clone));
      }
      return r;
    }).catch(() => caches.match(e.request))
  );
});
