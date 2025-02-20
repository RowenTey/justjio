self.addEventListener('push', event => {
    console.log('[Service Worker] Push notification received!');
    const payload = event.data.json();
    console.log(`[Service Worker] Push payload: "${payload}"`);

    const title = payload.title;
    const options = {
        body: payload.message,
    };

    event.waitUntil(self.registration.showNotification(title, options));
});