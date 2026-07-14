<script lang="ts">
  import { onMount, tick } from "svelte";
  import Auth from "./lib/Auth.svelte";
  import Map from "./lib/Map.svelte";
  import CreateMeetupForm from "./lib/CreateMeetupForm.svelte";
  import MeetupList from "./lib/MeetupList.svelte";

  // Master meetups data state
  let allMeetups = $state<any[]>([]);

  // State variables for map clicking & Lên kèo communication
  let mapComponent = $state<any>(null);
  let clickedLat = $state<number | null>(null);
  let clickedLng = $state<number | null>(null);
  let selectedAddressText = $state<string>("");

  // Temporary states for transactional Modal location selection
  let tempLat = $state<number | null>(null);
  let tempLng = $state<number | null>(null);
  let tempAddressText = $state<string>("");
  let isResolvingAddress = $state<boolean>(false);

  // Shared Filter and Geolocation states
  let selectedCity = $state<"all" | "HCM" | "HN">("all");
  let selectedDistance = $state<"all" | "5" | "10">("all");
  let userLat = $state<number | null>(null);
  let userLng = $state<number | null>(null);
  let isTrackingGPS = $state<boolean>(false);
  let gpsError = $state<boolean>(false);

  // Active navigation tab
  let activeTab = $state<"find" | "create" | "profile">("find");

  // Smart location selection mode
  let isSelectingLocation = $state<boolean>(false);

  // Modal visibility state
  let isMapModalOpen = $state<boolean>(false);

  // General API Health state
  let apiStatus: "online" | "offline" | "connecting" = $state("connecting");
  let apiMessage: string = $state("Đang ping server Go...");
  let apiCode: number | null = $state(null);
  let isChecking: boolean = $state(false);

  // Distance helper (Haversine)
  function calculateDistance(
    lat1: number,
    lon1: number,
    lat2: number,
    lon2: number,
  ): number {
    const R = 6371; // km
    const dLat = ((lat2 - lat1) * Math.PI) / 180;
    const dLon = ((lon2 - lon1) * Math.PI) / 180;
    const a =
      Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos((lat1 * Math.PI) / 180) *
        Math.cos((lat2 * Math.PI) / 180) *
        Math.sin(dLon / 2) *
        Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return R * c;
  }

  function getMeetupCity(meetup: any): "HCM" | "HN" {
    const distToHCM = calculateDistance(
      meetup.lat,
      meetup.lng,
      10.7769,
      106.7009,
    );
    const distToHN = calculateDistance(
      meetup.lat,
      meetup.lng,
      21.0285,
      105.8542,
    );
    return distToHN < distToHCM ? "HN" : "HCM";
  }

  // Svelte 5 reactive computed derived state
  let filteredMeetups = $derived.by(() => {
    return allMeetups.filter((meetup) => {
      const city = getMeetupCity(meetup);
      if (selectedCity !== "all" && city !== selectedCity) return false;
      if (selectedDistance !== "all" && userLat !== null && userLng !== null) {
        const dist = calculateDistance(
          userLat,
          userLng,
          meetup.lat,
          meetup.lng,
        );
        const limit = parseFloat(selectedDistance);
        if (dist > limit) return false;
      }
      return true;
    });
  });

  async function checkBackendHealth() {
    isChecking = true;
    apiStatus = "connecting";
    apiMessage = "Đang ping server...";
    apiCode = null;

    try {
      const res = await fetch("http://localhost:8080/api/health");
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
    } catch (err: any) {
      apiStatus = "offline";
      apiMessage = "Mất kết nối API";
      apiCode = null;
    } finally {
      isChecking = false;
    }
  }

  async function fetchAllMeetups() {
    try {
      const res = await fetch("http://localhost:8080/api/meetups");
      if (res.ok) {
        allMeetups = await res.json();
      }
    } catch (err) {
      console.error("Failed to fetch meetups from API:", err);
    }
  }

  async function reverseGeocodeTemp(targetLat: number, targetLng: number) {
    const token = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;
    if (!token) return;

    isResolvingAddress = true;
    const url = `https://api.mapbox.com/geocoding/v5/mapbox.places/${targetLng},${targetLat}.json?access_token=${token}&country=vn&limit=1&language=vi`;

    try {
      const res = await fetch(url);
      if (res.ok) {
        const data = await res.json();
        if (data.features && data.features.length > 0) {
          tempAddressText = data.features[0].place_name;
        } else {
          tempAddressText = `Tọa độ: ${targetLat.toFixed(5)}, ${targetLng.toFixed(5)}`;
        }
      } else {
        tempAddressText = `Tọa độ: ${targetLat.toFixed(5)}, ${targetLng.toFixed(5)}`;
      }
    } catch (err) {
      console.error("Temp reverse geocoding failed:", err);
      tempAddressText = `Tọa độ: ${targetLat.toFixed(5)}, ${targetLng.toFixed(5)}`;
    } finally {
      isResolvingAddress = false;
    }
  }

  function handleTempLocationSelect(lng: number, lat: number) {
    tempLat = lat;
    tempLng = lng;

    if (isSelectingLocation) {
      reverseGeocodeTemp(lat, lng);
    }
  }

  async function handleCreateSuccess() {
    clickedLat = null;
    clickedLng = null;
    selectedAddressText = "";
    isMapModalOpen = false;
    await fetchAllMeetups();
    activeTab = "find";
  }

  function handleSelectMeetup(meetup: any) {
    tempLat = meetup.lat;
    tempLng = meetup.lng;
    clickedLat = meetup.lat;
    clickedLng = meetup.lng;
    isSelectingLocation = false;
    isMapModalOpen = true;

    // Wait for Svelte modal DOM mount, then trigger jump
    tick().then(() => {
      if (mapComponent) {
        mapComponent.flyToMeetup(meetup);
      }
    });
  }

  function handleLocationInputClick() {
    openMapModal(true);
  }

  // Adjusted to load current form coordinates when opening select mode
  function openMapModal(selectingLocation = false) {
    isSelectingLocation = selectingLocation;
    if (selectingLocation) {
      tempLat = clickedLat;
      tempLng = clickedLng;
      tempAddressText = selectedAddressText;
    } else {
      tempLat = null;
      tempLng = null;
      tempAddressText = "";
    }
    isMapModalOpen = true;
  }

  function closeMapModal() {
    isMapModalOpen = false;
    isSelectingLocation = false;
    tempLat = null;
    tempLng = null;
    tempAddressText = "";
  }

  function handleBackdropClick(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (target.classList.contains("cartoon-modal-backdrop")) {
      closeMapModal();
    }
  }

  function handleBackdropKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape" || e.key === "Enter") {
      closeMapModal();
    }
  }

  function confirmLocationSelection() {
    clickedLat = tempLat;
    clickedLng = tempLng;
    selectedAddressText = tempAddressText;
    closeMapModal();
  }

  onMount(() => {
    checkBackendHealth();
    fetchAllMeetups();

    // Track User Geolocation
    if (navigator.geolocation) {
      isTrackingGPS = true;
      navigator.geolocation.getCurrentPosition(
        (position) => {
          userLat = position.coords.latitude;
          userLng = position.coords.longitude;
          isTrackingGPS = false;
        },
        (err) => {
          console.warn("GPS lookup failed:", err);
          gpsError = true;
          isTrackingGPS = false;
        },
        { enableHighAccuracy: true },
      );
    } else {
      gpsError = true;
    }
  });
</script>

<!-- Header / Navigation (Desktop) -->
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
      <span>Boardgame Hub</span>
    </div>

    <ul class="nav-links">
      <li>
        <button
          class="nav-link {activeTab === 'find' ? 'active' : ''}"
          onclick={() => (activeTab = "find")}
        >
          Tìm kèo
        </button>
      </li>
      <li>
        <button
          class="nav-link {activeTab === 'create' ? 'active' : ''}"
          onclick={() => (activeTab = "create")}
        >
          Lên kèo
        </button>
      </li>
      <li>
        <button
          class="nav-link {activeTab === 'profile' ? 'active' : ''}"
          onclick={() => (activeTab = "profile")}
        >
          Hồ sơ
        </button>
      </li>
    </ul>

    <div class="desktop-nav-btn">
      <button class="btn btn-primary" onclick={() => (activeTab = "profile")}>
        Tài khoản
      </button>
    </div>
  </div>
</header>

<main class="container" style="padding-top: 24px;">
  {#if activeTab === "find"}
    <!-- Tab Tìm Kèo -->
    <section
      id="map-discovery"
      style="text-align: center; padding-bottom: 20px;"
    >
      <!-- Premium Hero Header Card instead of static map -->
      <div
        class="cartoon-card"
        style="background-color: var(--pastel-cyan); padding: 30px 20px; text-align: center; margin-bottom: 32px; position: relative;"
      >
        <span
          class="badge-tag"
          style="background-color: #fff; margin-bottom: 12px;"
          >Bản Đồ Hội Nhóm 🗺️</span
        >
        <h2 style="font-size: 1.8rem; margin-bottom: 8px;">
          Khám phá các nhóm chơi quanh bạn
        </h2>
        <p
          style="font-weight: 600; color: var(--text-dark); max-width: 600px; margin: 0 auto 24px; line-height: 1.5; font-size: 0.95rem;"
        >
          Bạn muốn tìm các bàn chơi đang hoạt động trực quan địa lý? Hãy click
          mở bản đồ để quét các quân Meeple màu sắc!
        </p>
        <button
          class="btn btn-primary"
          style="font-size: 1.1rem; padding: 12px 32px;"
          onclick={() => openMapModal(false)}
        >
          🗺️ MỞ BẢN ĐỒ HỘI NHÓM
        </button>
      </div>

      <MeetupList
        meetups={filteredMeetups}
        {userLat}
        {userLng}
        bind:selectedCity
        bind:selectedDistance
        {isTrackingGPS}
        {gpsError}
        onSelectMeetup={handleSelectMeetup}
      />
    </section>
  {:else}
    <!-- Content header when in other tabs -->
    <section class="hero-section" style="padding: 24px 0 16px;">
      <span class="badge-tag">🎲 Boardgame Hub</span>
    </section>
  {/if}

  {#if activeTab === "create"}
    <!-- Tab Lên Kèo -->
    <section id="create-meetup-tab" style="padding-bottom: 40px;">
      <CreateMeetupForm
        bind:selectedLat={clickedLat}
        bind:selectedLng={clickedLng}
        bind:addressText={selectedAddressText}
        {userLat}
        {userLng}
        onLocationInputClick={handleLocationInputClick}
        onCreateSuccess={handleCreateSuccess}
      />
    </section>
  {/if}

  {#if activeTab === "profile"}
    <!-- Tab Hồ Sơ -->
    <section id="profile-tab" style="padding-bottom: 40px;">
      <h2 class="section-title">Hồ Sơ Của Bạn</h2>
      <Auth />

      <!-- Retro Console moved to debug tools section at bottom of profile -->
      <div style="margin-top: 40px;">
        <h3
          style="font-size: 1.1rem; color: var(--text-muted); font-weight: 700; margin-bottom: 12px;"
        >
          CÔNG CỤ PHÁT TRIỂN (API STATUS)
        </h3>
        <section class="retro-console" id="status" style="margin-top: 0;">
          <div class="console-screen">
            <div class="console-text-row">
              <span class="console-label">SERVER ADDR:</span>
              <span class="console-val">http://localhost:8080</span>
            </div>
            <div class="console-text-row">
              <span class="console-label">ENDPOINT:</span>
              <span class="console-val">/api/health</span>
            </div>
            <div class="console-text-row">
              <span class="console-label">STATUS CODE:</span>
              <span class="console-val"
                >{apiCode !== null ? apiCode : "---"}</span
              >
            </div>
            <div class="console-text-row">
              <span class="console-label">RESPONSE:</span>
              <span
                class="console-val {apiStatus === 'online'
                  ? 'online'
                  : apiStatus === 'offline'
                    ? 'offline'
                    : 'connecting'}"
              >
                {apiMessage}
              </span>
            </div>
          </div>

          <div class="console-controls">
            <div class="led-indicator">
              {#if apiStatus === "online"}
                <span class="led-dot led-online"></span>
                <span>ONLINE</span>
              {:else if apiStatus === "offline"}
                <span class="led-dot led-offline"></span>
                <span>OFFLINE</span>
              {:else}
                <span class="led-dot led-connecting"></span>
                <span>PINGING</span>
              {/if}
            </div>

            <div class="console-buttons">
              <button
                class="btn btn-success"
                style="padding: 8px 16px; font-size: 0.9rem;"
                onclick={checkBackendHealth}
                disabled={isChecking}
              >
                {isChecking ? "Pinging..." : "PING API"}
              </button>
            </div>
          </div>
        </section>
      </div>
    </section>
  {/if}
</main>

<!-- Mobile Bottom Navigation (Tab Bar) - Rebuilt to exactly 3 tabs -->
<nav class="mobile-nav-bar">
  <button
    class="mobile-nav-item {activeTab === 'create' ? 'active' : ''}"
    onclick={() => (activeTab = "create")}
  >
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2.5"
      stroke-linecap="round"
      stroke-linejoin="round"
    >
      <path d="M12 5v14M5 12h14" stroke-width="3" />
    </svg>
    <span>Lên kèo</span>
  </button>

  <button
    class="mobile-nav-item {activeTab === 'find' ? 'active' : ''}"
    onclick={() => (activeTab = "find")}
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
    class="mobile-nav-item {activeTab === 'profile' ? 'active' : ''}"
    onclick={() => (activeTab = "profile")}
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

<!-- Modal Map Backdrop Container -->
{#if isMapModalOpen}
  <div
    class="cartoon-modal-backdrop"
    onclick={handleBackdropClick}
    onkeydown={handleBackdropKeyDown}
    role="button"
    tabindex="0"
  >
    <div class="cartoon-card cartoon-modal-content">
      <div class="modal-header">
        <h3>
          {isSelectingLocation
            ? "🎯 Chọn Vị Trí Lên Kèo"
            : "🗺️ Bản Đồ Hội Nhóm"}
        </h3>
        <button class="btn btn-close-modal" onclick={closeMapModal}>✕</button>
      </div>

      <div class="modal-body">
        <Map
          bind:this={mapComponent}
          meetups={filteredMeetups}
          selectedLat={tempLat}
          selectedLng={tempLng}
          {isSelectingLocation}
          onLocationSelect={handleTempLocationSelect}
        />
      </div>

      <div class="modal-footer">
        {#if isSelectingLocation}
          <span
            class="modal-footer-tip"
            style="flex: 1; min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; padding-right: 8px;"
          >
            {#if tempLat && tempLng}
              📍 Địa chỉ: <strong
                >{isResolvingAddress
                  ? "Đang quét vị trí..."
                  : tempAddressText}</strong
              >
            {:else}
              👉 Click vào điểm bất kỳ trên bản đồ để chọn vị trí.
            {/if}
          </span>
          <button
            type="button"
            class="btn btn-primary btn-confirm-location"
            disabled={tempLat === null || isResolvingAddress}
            onclick={confirmLocationSelection}
          >
            Xác nhận vị trí 🎯
          </button>
        {:else}
          <span class="modal-footer-tip"
            >💡 Click quân cờ Meeple để xem chi tiết kèo.</span
          >
          <button
            type="button"
            class="btn btn-secondary"
            onclick={closeMapModal}>Đóng bản đồ</button
          >
        {/if}
      </div>
    </div>
  </div>
{/if}
