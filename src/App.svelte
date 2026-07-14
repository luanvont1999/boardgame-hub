<script lang="ts">
  import { onMount } from 'svelte';
  import { collection, onSnapshot, doc, setDoc } from 'firebase/firestore';
  import { db } from './lib/firebase';
  import { currentRoute, navigateToTab, isChildRoute, type RouteParams } from './lib/router';

  // Route components
  import FindRoute from './routes/FindRoute.svelte';
  import CreateRoute from './routes/CreateRoute.svelte';
  import ProfileRoute from './routes/ProfileRoute.svelte';
  import MapRoute from './routes/MapRoute.svelte';
  import FilterRoute from './routes/FilterRoute.svelte';
  import ChatRoute from './routes/ChatRoute.svelte';
  import ManageRoute from './routes/ManageRoute.svelte';

  // ── Global App State ──────────────────────────────────────────────────────
  let allMeetups = $state<any[]>([]);

  // Location state shared across CreateRoute ↔ MapRoute
  let createLat = $state<number | null>(null);
  let createLng = $state<number | null>(null);
  let createAddressText = $state<string>('');

  // GPS + Filter shared state
  let selectedCity = $state<'all' | 'HCM' | 'HN'>('all');
  let selectedDistance = $state<'all' | '5' | '10'>('all');
  let userLat = $state<number | null>(null);
  let userLng = $state<number | null>(null);
  let isTrackingGPS = $state<boolean>(false);
  let gpsError = $state<boolean>(false);

  // API health (profile page)
  let apiStatus = $state<'online' | 'offline' | 'connecting'>('connecting');
  let apiMessage = $state<string>('Đang ping server Go...');
  let apiCode = $state<number | null>(null);
  let isChecking = $state<boolean>(false);
  const API_BASE = import.meta.env.VITE_API_URL || '';

  // ── Derived State ─────────────────────────────────────────────────────────
  let route = $derived(currentRoute());
  let childRoute = $derived(isChildRoute());

  function calculateDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
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

  function getMeetupCity(meetup: any): 'HCM' | 'HN' {
    const dHCM = calculateDistance(meetup.lat, meetup.lng, 10.7769, 106.7009);
    const dHN = calculateDistance(meetup.lat, meetup.lng, 21.0285, 105.8542);
    return dHN < dHCM ? 'HN' : 'HCM';
  }

  let filteredMeetups = $derived.by(() =>
    allMeetups.filter((m) => {
      if (selectedCity !== 'all' && getMeetupCity(m) !== selectedCity) return false;
      if (selectedDistance !== 'all' && userLat !== null && userLng !== null) {
        const dist = calculateDistance(userLat, userLng, m.lat, m.lng);
        if (dist > parseFloat(selectedDistance)) return false;
      }
      return true;
    })
  );

  // ── Seed Data ─────────────────────────────────────────────────────────────
  const SEED_MEETUPS = [
    { id: '1', title: 'Hội Ma Sói Đêm Trăng Q1', game: 'Ultimate Werewolf', hostName: 'Minh Tuấn', hostUid: 'default-host-1', lat: 10.7769, lng: 106.7009, playersCount: 11, playersNeeded: 15, time: '2026-07-10T19:30', color: '#bca0f5' },
    { id: '2', title: 'Sân Chơi Mèo Nổ Q3', game: 'Exploding Kittens', hostName: 'Thanh Trúc', hostUid: 'default-host-2', lat: 10.7828, lng: 106.6896, playersCount: 4, playersNeeded: 5, time: '2026-07-11T15:00', color: '#ffa4b2' },
    { id: '3', title: 'CLB Cờ Tỷ Phú Bình Thạnh', game: 'Monopoly Deal', hostName: 'Khánh Huy', hostUid: 'default-host-3', lat: 10.7981, lng: 106.7051, playersCount: 3, playersNeeded: 6, time: '2026-07-12T18:00', color: '#ffe869' },
    { id: '4', title: 'Chiến Thần Catan Hoàn Kiếm', game: 'Settlers of Catan', hostName: 'Hoàng Lâm', hostUid: 'default-host-4', lat: 21.0285, lng: 105.8542, playersCount: 2, playersNeeded: 4, time: '2026-07-11T19:00', color: '#9ee3b2' },
    { id: '5', title: 'Hội Avalon Tây Hồ', game: 'Avalon', hostName: 'Thu Giang', hostUid: 'default-host-5', lat: 21.0588, lng: 105.8285, playersCount: 5, playersNeeded: 10, time: '2026-07-12T14:30', color: '#a4f0fd' },
  ];

  // ── Firestore Listener ────────────────────────────────────────────────────
  function listenToMeetupsRealtime() {
    onSnapshot(
      collection(db, 'meetups'),
      async (snapshot) => {
        if (snapshot.empty) {
          for (const seed of SEED_MEETUPS) await setDoc(doc(db, 'meetups', seed.id), seed);
        } else {
          allMeetups = snapshot.docs.map((d) => ({ id: d.id, ...d.data() }));
        }
      },
      (err) => console.error('[Firestore] meetups error:', err)
    );
  }

  // ── API Health ─────────────────────────────────────────────────────────────
  async function checkBackendHealth() {
    isChecking = true;
    apiStatus = 'connecting';
    apiMessage = 'Đang ping server...';
    apiCode = null;
    try {
      const res = await fetch(`${API_BASE}/api/health`);
      if (res.ok) {
        const data = await res.json();
        apiStatus = 'online';
        apiMessage = data.message || 'API hoạt động tốt!';
        apiCode = res.status;
      } else {
        apiStatus = 'offline';
        apiMessage = `Lỗi server: Code ${res.status}`;
        apiCode = res.status;
      }
    } catch {
      apiStatus = 'offline';
      apiMessage = 'Mất kết nối API';
    } finally {
      isChecking = false;
    }
  }

  // ── Lifecycle ─────────────────────────────────────────────────────────────
  onMount(() => {
    checkBackendHealth();
    listenToMeetupsRealtime();

    if (navigator.geolocation) {
      isTrackingGPS = true;
      navigator.geolocation.getCurrentPosition(
        (pos) => { userLat = pos.coords.latitude; userLng = pos.coords.longitude; isTrackingGPS = false; },
        (err) => { console.warn('GPS failed:', err); gpsError = true; isTrackingGPS = false; },
        { enableHighAccuracy: true }
      );
    } else {
      gpsError = true;
    }
  });

  // ── Filter Apply Handler ───────────────────────────────────────────────────
  function handleApplyFilter(city: 'all' | 'HCM' | 'HN', distance: 'all' | '5' | '10') {
    selectedCity = city;
    selectedDistance = distance;
  }
</script>

<!-- ── Navbar (Desktop) ──────────────────────────────────────────────────── -->
<header class="navbar">
  <div class="container navbar-container">
    <div class="logo">
      <svg class="logo-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <rect x="3" y="3" width="18" height="18" rx="2" ry="2" fill="currentColor"/>
        <circle cx="8" cy="8" r="1.5" fill="#fff"/>
        <circle cx="16" cy="16" r="1.5" fill="#fff"/>
        <circle cx="16" cy="8" r="1.5" fill="#fff"/>
        <circle cx="8" cy="16" r="1.5" fill="#fff"/>
        <circle cx="12" cy="12" r="1.5" fill="#fff"/>
      </svg>
      <span>Boardgame Luna</span>
    </div>

    <ul class="nav-links">
      <li><button class="nav-link {route.name === 'find' ? 'active' : ''}" onclick={() => navigateToTab('find')}>Tìm kèo</button></li>
      <li><button class="nav-link {route.name === 'create' ? 'active' : ''}" onclick={() => navigateToTab('create')}>Lên kèo</button></li>
      <li><button class="nav-link {route.name === 'profile' ? 'active' : ''}" onclick={() => navigateToTab('profile')}>Hồ sơ</button></li>
    </ul>

    <div class="desktop-nav-btn">
      <button class="btn btn-primary" onclick={() => navigateToTab('profile')}>Tài khoản</button>
    </div>
  </div>
</header>

<!-- ── Main Content ──────────────────────────────────────────────────────── -->
<main class="container" style="padding-top: 24px;">

  {#if route.name === 'find'}
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

  {:else if route.name === 'create'}
    <CreateRoute
      bind:selectedLat={createLat}
      bind:selectedLng={createLng}
      bind:addressText={createAddressText}
      {userLat}
      {userLng}
    />

  {:else if route.name === 'profile'}
    <ProfileRoute
      {apiStatus}
      {apiMessage}
      {apiCode}
      apiBase={API_BASE}
      {isChecking}
      onCheckHealth={checkBackendHealth}
    />

  {:else if route.name === 'map'}
    <MapRoute
      meetups={allMeetups}
      bind:selectedLat={createLat}
      bind:selectedLng={createLng}
      bind:addressText={createAddressText}
      mode={route.mode}
    />

  {:else if route.name === 'filter'}
    <FilterRoute
      {selectedCity}
      {selectedDistance}
      {userLat}
      {isTrackingGPS}
      {gpsError}
      onApply={handleApplyFilter}
    />

  {:else if route.name === 'chat'}
    <ChatRoute meetup={route.meetup} />

  {:else if route.name === 'manage'}
    <ManageRoute meetup={route.meetup} />
  {/if}

</main>

<!-- ── Mobile Bottom Tab Bar (hidden for child routes) ──────────────────── -->
{#if !childRoute}
  <nav class="mobile-nav-bar">
    <button class="mobile-nav-item {route.name === 'create' ? 'active' : ''}" onclick={() => navigateToTab('create')}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 5v14M5 12h14" stroke-width="3"/>
      </svg>
      <span>Lên kèo</span>
    </button>

    <button class="mobile-nav-item {route.name === 'find' ? 'active' : ''}" onclick={() => navigateToTab('find')}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="10"/>
        <polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"/>
      </svg>
      <span>Tìm kèo</span>
    </button>

    <button class="mobile-nav-item {route.name === 'profile' ? 'active' : ''}" onclick={() => navigateToTab('profile')}>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
        <circle cx="12" cy="7" r="4"/>
      </svg>
      <span>Hồ sơ</span>
    </button>
  </nav>
{/if}
