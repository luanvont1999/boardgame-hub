// public/firebase-messaging-sw.js
importScripts('https://www.gstatic.com/firebasejs/10.8.0/firebase-app-compat.js');
importScripts('https://www.gstatic.com/firebasejs/10.8.0/firebase-messaging-compat.js');

firebase.initializeApp({
  apiKey: "AIzaSyATafzB15sNKoZp1dZ5_e0enrwpQr99-Vc",
  authDomain: "boardgame-hub-7f7a2.firebaseapp.com",
  projectId: "boardgame-hub-7f7a2",
  storageBucket: "boardgame-hub-7f7a2.firebasestorage.app",
  messagingSenderId: "202469575377",
  appId: "1:202469575377:web:82470f41441d81ebbc53f0"
});

const messaging = firebase.messaging();

// Lắng nghe thông báo khi ứng dụng chạy ngầm / đóng
messaging.onBackgroundMessage((payload) => {
  console.log('[firebase-messaging-sw.js] Nhận thông báo ngầm:', payload);
  
  const notificationTitle = payload.notification?.title || 'Boardgame Luna 🎲';
  const notificationOptions = {
    body: payload.notification?.body || 'Bạn có thông báo mới!',
    icon: '/boardgame_pwa_icon_1784017090071.png',
    badge: '/boardgame_pwa_icon_1784017090071.png',
    data: payload.data || {}
  };

  self.registration.showNotification(notificationTitle, notificationOptions);
});
