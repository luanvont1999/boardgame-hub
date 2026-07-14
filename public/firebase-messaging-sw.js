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
  
  // Nếu payload có trường 'notification', Firebase SDK sẽ tự động hiển thị thông báo hệ thống
  // dựa trên cấu hình gửi từ Server (chứa icon, badge, v.v.).
  // Việc tự gọi showNotification một lần nữa ở đây sẽ gây xung đột và kích hoạt thông báo mặc định 
  // "Trang web này được cập nhật trong trang nền" của trình duyệt để cảnh báo.
  if (!payload.notification) {
    const notificationTitle = payload.data?.title || 'Boardgame Luna 🎲';
    const notificationOptions = {
      body: payload.data?.body || 'Bạn có thông báo mới!',
      icon: '/boardgame_pwa_icon_1784017090071.png',
      badge: '/boardgame_pwa_icon_1784017090071.png',
      data: payload.data || {}
    };

    self.registration.showNotification(notificationTitle, notificationOptions);
  }
});

// Xử lý sự kiện click vào thông báo đẩy để chuyển hướng
self.addEventListener('notificationclick', (event) => {
  event.notification.close();

  // Trích xuất link điều hướng từ data payload
  let clickAction = '/';
  if (event.notification.data) {
    clickAction = event.notification.data.clickAction || clickAction;
    if (!event.notification.data.clickAction && event.notification.data.FCM_MSG) {
      const fcmMsg = event.notification.data.FCM_MSG;
      if (fcmMsg.notification && fcmMsg.notification.click_action) {
        clickAction = fcmMsg.notification.click_action;
      }
    }
  }

  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then((windowClients) => {
      // Nếu có tab app đang mở, focus vào tab đó và gửi tin nhắn điều hướng route
      for (let i = 0; i < windowClients.length; i++) {
        const client = windowClients[i];
        if ('focus' in client) {
          client.focus();
          client.postMessage({ type: 'NAVIGATE_ROUTE', url: clickAction });
          return;
        }
      }
      // Nếu chưa có tab nào mở, mở tab mới với đường dẫn clickAction
      if (clients.openWindow) {
        return clients.openWindow(clickAction);
      }
    })
  );
});
