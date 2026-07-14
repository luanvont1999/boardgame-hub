<script lang="ts">
  import { onMount } from "svelte";
  import { collection, onSnapshot, doc, setDoc } from "firebase/firestore";
  import { db, auth } from "./lib/firebase";
  import { onAuthStateChanged, type User } from "firebase/auth";
  import {
    currentRoute,
    navigateToTab,
    isChildRoute,
    type RouteParams,
  } from "./lib/router.svelte";
  import { initNotifications } from "./lib/notificationService";

  // Route components
  import FindRoute from "./routes/FindRoute.svelte";
  import CreateRoute from "./routes/CreateRoute.svelte";
  import ProfileRoute from "./routes/ProfileRoute.svelte";
  import MapRoute from "./routes/MapRoute.svelte";
  import FilterRoute from "./routes/FilterRoute.svelte";
  import ChatRoute from "./routes/ChatRoute.svelte";
  import ManageRoute from "./routes/ManageRoute.svelte";

  // ── Global App State ──────────────────────────────────────────────────────
  let allMeetups = $state<any[]>([]);
  let currentUser = $state<User | null>(auth.currentUser);

  // Toasts state
  interface Toast {
    id: string;
    message: string;
    type: "info" | "success" | "warning";
  }
  let toasts = $state<Toast[]>([]);

  function addToast(message: string, type: "info" | "success" | "warning" = "info") {
    const id = Math.random().toString(36).substring(2, 9);
    toasts = [...toasts, { id, message, type }];
    setTimeout(() => {
      toasts = toasts.filter((t) => t.id !== id);
    }, 4000);
  }

  // Location state shared across CreateRoute ↔ MapRoute
  let createLat = $state<number | null>(null);
  let createLng = $state<number | null>(null);
  let createAddressText = $state<string>("");

  // GPS + Filter shared state
  let selectedCity = $state<"all" | "HCM" | "HN">("all");
  let selectedDistance = $state<"all" | "5" | "10">("all");
  let userLat = $state<number | null>(null);
  let userLng = $state<number | null>(null);
  let isTrackingGPS = $state<boolean>(false);
  let gpsError = $state<boolean>(false);

  // API health (profile page)
  let apiStatus = $state<"online" | "offline" | "connecting">("connecting");
  let apiMessage = $state<string>("Đang ping server Go...");
  let apiCode = $state<number | null>(null);
  let isChecking = $state<boolean>(false);
  const API_BASE = import.meta.env.VITE_API_URL || "";

  // PWA Install state
  let deferredPrompt = $state<any>(null);
  let showInstallBanner = $state(true);
  let showIOSInstallInstructions = $state(false);
  let isIOS = $state(false);
  let isStandalone = $state(false);

  // ── Derived State ─────────────────────────────────────────────────────────
  let route = $derived(currentRoute());
  let childRoute = $derived(isChildRoute());

  // Tính tổng số request đang chờ duyệt đối với các kèo do mình làm host
  let totalPendingRequests = $derived.by(() => {
    if (!currentUser) return 0;
    return allMeetups
      .filter((m) => m.hostUid === currentUser.uid || m.host_uid === currentUser.uid)
      .reduce((sum, m) => sum + (Array.isArray(m.pendingUids) ? m.pendingUids.length : 0), 0);
  });

  function calculateDistance(
    lat1: number,
    lon1: number,
    lat2: number,
    lon2: number,
  ): number {
    const R = 6371;
    const dLat = ((lat2 - lat1) * Math.PI) / 180;
    const dLon = ((lon2 - lon1) * Math.PI) / 180;
    const a =
      Math.sin(dLat / 2) ** 2 +
      Math.cos((lat1 * Math.PI) / 180) *
        Math.cos((lat2 * Math.PI) / 180) *
        Math.sin(dLon / 2) ** 2;
    return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
  }

  function getMeetupCity(meetup: any): "HCM" | "HN" {
    const dHCM = calculateDistance(meetup.lat, meetup.lng, 10.7769, 106.7009);
    const dHN = calculateDistance(meetup.lat, meetup.lng, 21.0285, 105.8542);
    return dHN < dHCM ? "HN" : "HCM";
  }

  let filteredMeetups = $derived.by(() =>
    allMeetups.filter((m) => {
      if (selectedCity !== "all" && getMeetupCity(m) !== selectedCity)
        return false;
      if (selectedDistance !== "all" && userLat !== null && userLng !== null) {
        const dist = calculateDistance(userLat, userLng, m.lat, m.lng);
        if (dist > parseFloat(selectedDistance)) return false;
      }
      return true;
    }),
  );

  // ── Seed Data ─────────────────────────────────────────────────────────────
  const SEED_MEETUPS = [
    {
      id: "1",
      title: "Hội Ma Sói Đêm Trăng Q1",
      game: "Ultimate Werewolf",
      hostName: "Minh Tuấn",
      hostUid: "default-host-1",
      lat: 10.7769,
      lng: 106.7009,
      playersCount: 11,
      playersNeeded: 15,
      time: "2026-07-10T19:30",
      color: "#bca0f5",
    },
    {
      id: "2",
      title: "Sân Chơi Mèo Nổ Q3",
      game: "Exploding Kittens",
      hostName: "Thanh Trúc",
      hostUid: "default-host-2",
      lat: 10.7828,
      lng: 106.6896,
      playersCount: 4,
      playersNeeded: 5,
      time: "2026-07-11T15:00",
      color: "#ffa4b2",
    },
    {
      id: "3",
      title: "CLB Cờ Tỷ Phú Bình Thạnh",
      game: "Monopoly Deal",
      hostName: "Khánh Huy",
      hostUid: "default-host-3",
      lat: 10.7981,
      lng: 106.7051,
      playersCount: 3,
      playersNeeded: 6,
      time: "2026-07-12T18:00",
      color: "#ffe869",
    },
    {
      id: "4",
      title: "Chiến Thần Catan Hoàn Kiếm",
      game: "Settlers of Catan",
      hostName: "Hoàng Lâm",
      hostUid: "default-host-4",
      lat: 21.0285,
      lng: 105.8542,
      playersCount: 2,
      playersNeeded: 4,
      time: "2026-07-11T19:00",
      color: "#9ee3b2",
    },
    {
      id: "5",
      title: "Hội Avalon Tây Hồ",
      game: "Avalon",
      hostName: "Thu Giang",
      hostUid: "default-host-5",
      lat: 21.0588,
      lng: 105.8285,
      playersCount: 5,
      playersNeeded: 10,
      time: "2026-07-12T14:30",
      color: "#a4f0fd",
    },
  ];

  // ── Firestore Listener ────────────────────────────────────────────────────
  function listenToMeetupsRealtime() {
    onSnapshot(
      collection(db, "meetups"),
      async (snapshot) => {
        if (snapshot.empty) {
          for (const seed of SEED_MEETUPS)
            await setDoc(doc(db, "meetups", seed.id), seed);
        } else {
          allMeetups = snapshot.docs.map((d) => ({ id: d.id, ...d.data() }));
        }
      },
      (err) => console.error("[Firestore] meetups error:", err),
    );
  }

  // ── API Health ─────────────────────────────────────────────────────────────
  async function checkBackendHealth() {
    isChecking = true;
    apiStatus = "connecting";
    apiMessage = "Đang ping server...";
    apiCode = null;
    try {
      const res = await fetch(`${API_BASE}/api/health`);
      if (res.ok) {
        const data = await res.json();
        apiStatus = "online";
        apiMessage = data.message || "API hoạt động tốt!";
        apiCode = res.status;
      } else {
        apiStatus = "offline";
        apiMessage = `Lỗi server: Code ${res.status}`;
        apiCode = res.status;
      }
    } catch {
      apiStatus = "offline";
      apiMessage = "Mất kết nối API";
    } finally {
      isChecking = false;
    }
  }

  // ── Lifecycle ─────────────────────────────────────────────────────────────
  let prevPendingCount = 0;
  let hasLoadedInitialMeetups = false;

  $effect(() => {
    // Đánh dấu đã load xong dữ liệu ban đầu khi allMeetups không còn rỗng
    if (allMeetups.length > 0 && !hasLoadedInitialMeetups) {
      setTimeout(() => {
        prevPendingCount = totalPendingRequests;
        hasLoadedInitialMeetups = true;
      }, 500);
      return;
    }

    if (hasLoadedInitialMeetups && totalPendingRequests > prevPendingCount) {
      const diff = totalPendingRequests - prevPendingCount;
      addToast(`🔔 Bạn có ${diff} yêu cầu tham gia kèo mới đang chờ duyệt!`, "info");
    }

    if (hasLoadedInitialMeetups) {
      prevPendingCount = totalPendingRequests;
    }
  });

  function handleDeepLink(urlStr: string) {
    try {
      const url = new URL(urlStr, window.location.origin);
      const routeParam = url.searchParams.get('route');
      if (routeParam === 'manage') {
        const meetupId = url.searchParams.get('meetupId');
        if (meetupId && allMeetups.length > 0) {
          const matchedMeetup = allMeetups.find(m => m.id === meetupId);
          if (matchedMeetup) {
            navigate({ name: 'manage', params: { meetup: matchedMeetup } });
          }
        }
      } else if (routeParam === 'profile') {
        navigateToTab('profile');
      }
    } catch (e) {
      console.error("Lỗi parse deep link:", e);
    }
  }

  $effect(() => {
    if (allMeetups.length > 0) {
      const params = new URLSearchParams(window.location.search);
      const routeParam = params.get('route');
      if (routeParam === 'manage') {
        const meetupId = params.get('meetupId');
        const matchedMeetup = allMeetups.find(m => m.id === meetupId);
        if (matchedMeetup) {
          window.history.replaceState({}, document.title, window.location.pathname);
          navigate({ name: 'manage', params: { meetup: matchedMeetup } });
        }
      } else if (routeParam === 'profile') {
        window.history.replaceState({}, document.title, window.location.pathname);
        navigateToTab('profile');
      }
    }
  });

  onMount(() => {
    checkBackendHealth();
    listenToMeetupsRealtime();

    // Check PWA & iOS environment
    isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent) && !(window as any).MSStream;
    isStandalone = window.matchMedia('(display-mode: standalone)').matches || (navigator as any).standalone;

    window.addEventListener('beforeinstallprompt', (e) => {
      e.preventDefault();
      deferredPrompt = e;
    });

    window.addEventListener('appinstalled', () => {
      deferredPrompt = null;
      addToast("🎉 Cài đặt ứng dụng thành công!", "success");
    });

    if ('serviceWorker' in navigator) {
      navigator.serviceWorker.addEventListener('message', (event) => {
        if (event.data && event.data.type === 'NAVIGATE_ROUTE') {
          handleDeepLink(event.data.url);
        }
      });
    }

    const unsubAuth = onAuthStateChanged(auth, (user) => {
      currentUser = user;
      if (user) {
        initNotifications(user.uid, (payload) => {
          const title = payload.notification?.title || "Thông báo mới";
          const body = payload.notification?.body || "";
          addToast(`🔔 ${title}: ${body}`, "info");

          // Hiển thị thông báo hệ thống (system notification banner) ở chế độ Foreground
          if (Notification.permission === 'granted') {
            const clickAction = payload.data?.clickAction || '/';
            const notification = new Notification(title, {
              body: body,
              icon: '/boardgame_pwa_icon_1784017090071.png',
              tag: 'foreground-push'
            });
            notification.onclick = () => {
              window.focus();
              handleDeepLink(clickAction);
              notification.close();
            };
          }
        });
      }
    });

    if (navigator.geolocation) {
      isTrackingGPS = true;
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          userLat = pos.coords.latitude;
          userLng = pos.coords.longitude;
          isTrackingGPS = false;
        },
        (err) => {
          console.warn("GPS failed:", err);
          gpsError = true;
          isTrackingGPS = false;
        },
        { enableHighAccuracy: true },
      );
    } else {
      gpsError = true;
    }

    return () => {
      unsubAuth();
    };
  });

  // ── Filter Apply Handler ───────────────────────────────────────────────────
  function handleApplyFilter(
    city: "all" | "HCM" | "HN",
    distance: "all" | "5" | "10",
  ) {
    selectedCity = city;
    selectedDistance = distance;
  }

  // ── PWA Install Trigger ──────────────────────────────────────────────────
  async function triggerPWAInstall() {
    if (isIOS && !isStandalone) {
      showIOSInstallInstructions = true;
      return;
    }
    
    if (!deferredPrompt) {
      addToast("Trình duyệt không hỗ trợ cài đặt tự động hoặc ứng dụng đã được tải về. Vui lòng thêm thủ công từ menu trình duyệt!", "info");
      return;
    }

    deferredPrompt.prompt();
    const { outcome } = await deferredPrompt.userChoice;
    console.log(`Lựa chọn cài đặt PWA: ${outcome}`);
    if (outcome === 'accepted') {
      deferredPrompt = null;
    }
  }
</script>

<!-- ── Navbar (Desktop) ──────────────────────────────────────────────────── -->
<header class="navbar">
  <div class="container navbar-container">
    <div class="logo">
      <svg
        class="logo-icon"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <rect
          x="3"
          y="3"
          width="18"
          height="18"
          rx="2"
          ry="2"
          fill="currentColor"
        />
        <circle cx="8" cy="8" r="1.5" fill="#fff" />
        <circle cx="16" cy="16" r="1.5" fill="#fff" />
        <circle cx="16" cy="8" r="1.5" fill="#fff" />
        <circle cx="8" cy="16" r="1.5" fill="#fff" />
        <circle cx="12" cy="12" r="1.5" fill="#fff" />
      </svg>
      <span>Boardgame Luna</span>
    </div>

    <ul class="nav-links">
      <li>
        <button
          class="nav-link {route.name === 'find' ? 'active' : ''}"
          onclick={() => navigateToTab("find")}>Tìm kèo</button
        >
      </li>
      <li>
        <button
          class="nav-link {route.name === 'create' ? 'active' : ''}"
          onclick={() => navigateToTab("create")}
        >
          Lên kèo
          {#if totalPendingRequests > 0}
            <span class="nav-badge">{totalPendingRequests}</span>
          {/if}
        </button>
      </li>
      <li>
        <button
          class="nav-link {route.name === 'profile' ? 'active' : ''}"
          onclick={() => navigateToTab("profile")}>Hồ sơ</button
        >
      </li>
    </ul>

    <div class="desktop-nav-btn" style="display: flex; gap: 10px;">
      {#if deferredPrompt || (isIOS && !isStandalone)}
        <button class="btn btn-primary" style="background-color: var(--pastel-yellow, #ffe869) !important; color: #1e1e24 !important;" onclick={triggerPWAInstall}>
          Tải App 📲
        </button>
      {/if}
      <button class="btn btn-primary" onclick={() => navigateToTab("profile")}
        >Tài khoản</button
      >
    </div>
  </div>
</header>

<!-- ── Main Content ──────────────────────────────────────────────────────── -->
{#if (deferredPrompt || (isIOS && !isStandalone)) && showInstallBanner}
  <div class="container" style="margin-top: 15px; margin-bottom: -10px;">
    <div class="cartoon-card install-banner">
      <div class="install-banner-content">
        <span class="install-banner-icon">📲</span>
        <div class="install-banner-text">
          <strong>Cài đặt Boardgame Luna!</strong>
          <span>Tải app về màn hình chính để nhận thông báo đẩy nhanh hơn.</span>
        </div>
      </div>
      <div class="install-banner-actions">
        <button class="btn btn-success install-banner-btn" onclick={triggerPWAInstall}>Cài đặt</button>
        <button class="btn btn-secondary install-banner-close" onclick={() => showInstallBanner = false}>Lúc khác</button>
      </div>
    </div>
  </div>
{/if}

<main class="container">
  {#if route.name === "find"}
    <FindRoute
      meetups={allMeetups}
      {filteredMeetups}
      bind:selectedCity
      bind:selectedDistance
      {userLat}
      {userLng}
      {isTrackingGPS}
      {gpsError}
    />
  {:else if route.name === "create"}
    <CreateRoute
      meetups={allMeetups}
      bind:selectedLat={createLat}
      bind:selectedLng={createLng}
      bind:addressText={createAddressText}
      {userLat}
      {userLng}
    />
  {:else if route.name === "profile"}
    <ProfileRoute
      {apiStatus}
      {apiMessage}
      {apiCode}
      apiBase={API_BASE}
      {isChecking}
      onCheckHealth={checkBackendHealth}
      {addToast}
    />
  {:else if route.name === "map"}
    <MapRoute
      meetups={allMeetups}
      bind:selectedLat={createLat}
      bind:selectedLng={createLng}
      bind:addressText={createAddressText}
      mode={route.mode}
    />
  {:else if route.name === "filter"}
    <FilterRoute
      {selectedCity}
      {selectedDistance}
      {userLat}
      {isTrackingGPS}
      {gpsError}
      onApply={handleApplyFilter}
    />
  {:else if route.name === "chat"}
    <ChatRoute meetup={route.meetup} />
  {:else if route.name === "manage"}
    <ManageRoute meetup={route.meetup} />
  {/if}
</main>

<!-- Toast Container -->
<div class="toast-container">
  {#each toasts as toast (toast.id)}
    <div class="cartoon-card toast-item {toast.type}">
      <span class="toast-message">{toast.message}</span>
      <button class="toast-close-btn" onclick={() => toasts = toasts.filter((t) => t.id !== toast.id)}>✕</button>
    </div>
  {/each}
</div>

<!-- ── Mobile Bottom Tab Bar (hidden for child routes) ──────────────────── -->
{#if !childRoute}
  <nav class="mobile-nav-bar">
    <button
      class="mobile-nav-item {route.name === 'create' ? 'active' : ''}"
      onclick={() => navigateToTab("create")}
    >
      <div style="position: relative; display: inline-block;">
        <svg
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          style="width: 24px; height: 24px;"
        >
          <path d="M12 5v14M5 12h14" stroke-width="3" />
        </svg>
        {#if totalPendingRequests > 0}
          <span class="mobile-badge-count">{totalPendingRequests}</span>
        {/if}
      </div>
      <span>Lên kèo</span>
    </button>

    <button
      class="mobile-nav-item {route.name === 'find' ? 'active' : ''}"
      onclick={() => navigateToTab("find")}
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <circle cx="12" cy="12" r="10" />
        <polygon
          points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"
        />
      </svg>
      <span>Tìm kèo</span>
    </button>

    <button
      class="mobile-nav-item {route.name === 'profile' ? 'active' : ''}"
      onclick={() => navigateToTab("profile")}
    >
      <svg
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
        <circle cx="12" cy="7" r="4" />
      </svg>
      <span>Hồ sơ</span>
    </button>
  </nav>
{/if}

<style>
  .nav-badge {
    background-color: #ef4444;
    color: white;
    font-size: 0.75rem;
    font-weight: 800;
    padding: 2px 6px;
    border-radius: 10px;
    border: 2px solid #1e1e24;
    margin-left: 6px;
    vertical-align: middle;
    display: inline-block;
  }
  .mobile-badge-count {
    position: absolute;
    top: -5px;
    right: -10px;
    background-color: #ef4444;
    color: white;
    font-size: 0.7rem;
    font-weight: 800;
    padding: 1px 5px;
    border-radius: 10px;
    border: 1.5px solid #1e1e24;
    line-height: 1;
  }
  .toast-container {
    position: fixed;
    top: 24px;
    right: 24px;
    z-index: 99999;
    display: flex;
    flex-direction: column;
    gap: 12px;
    max-width: 340px;
    width: calc(100% - 48px);
  }
  .toast-item {
    background-color: #fffdfb;
    border: 3px solid #1e1e24;
    box-shadow: 4px 4px 0px #1e1e24;
    padding: 14px 18px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    animation: toast-slide-in 0.25s cubic-bezier(0.175, 0.885, 0.32, 1.275) forwards;
  }
  .toast-item.info {
    border-left: 10px solid var(--pastel-blue, #a4f0fd);
  }
  .toast-message {
    font-weight: 700;
    font-size: 0.95rem;
    color: #1e1e24;
    line-height: 1.4;
  }
  .toast-close-btn {
    background: none;
    border: none;
    font-size: 1.1rem;
    font-weight: 800;
    cursor: pointer;
    color: #1e1e24;
    padding: 0 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  @keyframes toast-slide-in {
    from {
      transform: translateY(-20px) scale(0.9);
      opacity: 0;
    }
    to {
      transform: translateY(0) scale(1);
      opacity: 1;
    }
  }
  @media (max-width: 768px) {
    .toast-container {
      top: auto;
      bottom: 90px;
      right: 24px;
    }
  }

  /* PWA Install Banner & Modal Styles */
  .install-banner {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background-color: var(--pastel-blue, #a4f0fd);
    padding: 12px 20px;
    border-radius: 12px;
    border: 3px solid #1e1e24;
    box-shadow: 4px 4px 0px #1e1e24;
    gap: 15px;
    flex-wrap: wrap;
  }
  .install-banner-content {
    display: flex;
    align-items: center;
    gap: 12px;
  }
  .install-banner-icon {
    font-size: 2rem;
  }
  .install-banner-text {
    text-align: left;
  }
  .install-banner-text strong {
    font-size: 1rem;
    color: #1e1e24;
    display: block;
    margin-bottom: 2px;
  }
  .install-banner-text span {
    font-size: 0.8rem;
    color: #1e1e24;
  }
  .install-banner-actions {
    display: flex;
    gap: 10px;
  }
  .install-banner-btn {
    padding: 6px 14px !important;
    font-size: 0.85rem !important;
  }
  .install-banner-close {
    padding: 6px 14px !important;
    font-size: 0.85rem !important;
    background-color: white !important;
  }
  .ios-modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background-color: rgba(30, 30, 36, 0.6);
    z-index: 100000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
  }
  .ios-instructions-modal {
    background-color: #fffdfb;
    max-width: 360px;
    width: 100%;
    padding: 24px;
    border-radius: 12px;
    border: 3px solid #1e1e24;
    box-shadow: 6px 6px 0px #1e1e24;
  }
  @media (max-width: 768px) {
    .install-banner {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;
    }
    .install-banner-actions {
      justify-content: flex-end;
    }
  }
</style>

<!-- iOS Install Modal -->
{#if showIOSInstallInstructions}
  <div class="ios-modal-backdrop">
    <div class="cartoon-card ios-instructions-modal">
      <div style="display: flex; justify-content: space-between; align-items: center; border-bottom: 3px solid #1e1e24; padding-bottom: 12px; margin-bottom: 15px;">
        <h3 style="margin: 0; font-size: 1.2rem; font-weight: 800; color: #1e1e24;">Cài đặt trên iPhone/iPad 📲</h3>
        <button style="background: none; border: none; font-size: 1.5rem; font-weight: 800; cursor: pointer; color: #1e1e24;" onclick={() => showIOSInstallInstructions = false}>✕</button>
      </div>
      <p style="font-size: 0.95rem; line-height: 1.5; color: #1e1e24; margin-bottom: 15px; text-align: left;">
        Trình duyệt Safari trên iOS không hỗ trợ tự động cài đặt. Hãy làm theo 3 bước đơn giản:
      </p>
      <ol style="font-size: 0.9rem; line-height: 1.6; color: #1e1e24; padding-left: 20px; margin-bottom: 20px; text-align: left;">
        <li style="margin-bottom: 8px;">Nhấn vào biểu tượng <strong>Chia sẻ (Share)</strong> <span style="font-size: 1.1rem; vertical-align: middle;">📤</span> trên Safari.</li>
        <li style="margin-bottom: 8px;">Cuộn xuống và chọn <strong>Thêm vào MH chính (Add to Home Screen)</strong> <span style="font-size: 1.1rem; vertical-align: middle;">➕</span>.</li>
        <li>Bấm <strong>Thêm (Add)</strong> ở góc trên bên phải để hoàn tất!</li>
      </ol>
      <button class="btn btn-primary" style="width: 100%; padding: 10px;" onclick={() => showIOSInstallInstructions = false}>Đã hiểu 👍</button>
    </div>
  </div>
{/if}
