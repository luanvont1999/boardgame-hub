<script lang="ts">
  import { onMount, onDestroy, tick } from "svelte";
  import mapboxgl from "mapbox-gl";

  interface Meetup {
    id: string;
    title: string;
    game: string;
    host_name?: string;
    hostName?: string;
    host_uid?: string;
    hostUid?: string;
    lat: number;
    lng: number;
    players_count?: number;
    playersCount?: number;
    players_needed?: number;
    playersNeeded?: number;
    time: string;
    color: string;
  }

  interface PresetVenue {
    name: string;
    address: string;
    lat: number;
    lng: number;
    distance?: number | null;
  }

  // Svelte 5 Props
  interface Props {
    meetups: Meetup[];
    presetVenues?: PresetVenue[];
    selectedLat: number | null;
    selectedLng: number | null;
    mode?: "discover" | "select";
    isSelectingLocation?: boolean;
    onLocationSelect?: (lng: number, lat: number) => void;
  }
  let {
    meetups,
    presetVenues = [],
    selectedLat,
    selectedLng,
    mode = "discover",
    isSelectingLocation = false,
    onLocationSelect,
  }: Props = $props();

  // Svelte 5 state variables
  let mapContainer = $state<HTMLDivElement | null>(null);
  let mapInstance = $state<mapboxgl.Map | null>(null);
  let customToken = $state<string>("");
  let isMapInitialized = $state<boolean>(false);
  let isTokenMissing = $state<boolean>(false);

  // Tracking markers
  let markersList: mapboxgl.Marker[] = [];
  let markersMap = new Map<string, mapboxgl.Marker>();
  let tempMarker: mapboxgl.Marker | null = null;

  function getMapboxToken(): string {
    return import.meta.env.VITE_MAPBOX_ACCESS_TOKEN || "";
  }

  function clearMarkers() {
    markersList.forEach((m) => m.remove());
    markersList = [];
    markersMap.clear();
  }

  function renderMarkers() {
    const map = mapInstance;
    if (!map) return;
    clearMarkers();

    if (mode === "select" && presetVenues.length > 0) {
      presetVenues.forEach((venue, index) => {
        const el = document.createElement("div");
        el.className = "custom-venue-marker";
        el.style.backgroundColor = "#ffe869"; // pastel yellow
        el.style.border = "2px solid #1e1e24";
        el.style.boxShadow = "2.5px 2.5px 0 #1e1e24";
        el.style.borderRadius = "50%";
        el.style.width = "30px";
        el.style.height = "30px";
        el.style.display = "flex";
        el.style.alignItems = "center";
        el.style.justifyContent = "center";
        el.style.cursor = "pointer";

        el.innerHTML = `
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="#1e1e24" stroke-width="2.2">
            <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
            <polyline points="9 22 9 12 15 12 15 22"/>
          </svg>
        `;

        const distText = venue.distance !== undefined && venue.distance !== null 
          ? ` (cách ${venue.distance < 1 ? `${Math.round(venue.distance * 1000)}m` : `${venue.distance.toFixed(1)}km`})`
          : '';

        const popupHTML = `
          <div class="cartoon-popup">
            <h4>${venue.name}</h4>
            <p>📍 ${venue.address}${distText}</p>
          </div>
        `;

        const popup = new mapboxgl.Popup({
          offset: 20,
          closeButton: false,
        }).setHTML(popupHTML);

        el.addEventListener("click", (e) => {
          e.stopPropagation();
          showTempMarker(venue.lng, venue.lat);
          if (onLocationSelect) {
            onLocationSelect(venue.lng, venue.lat);
          }
        });

        const marker = new mapboxgl.Marker(el)
          .setLngLat([venue.lng, venue.lat])
          .setPopup(popup)
          .addTo(map);

        markersList.push(marker);
      });
      return;
    }

    meetups.forEach((meetup) => {
      const el = document.createElement("div");
      el.className = "custom-meeple-marker";
      el.style.backgroundColor = meetup.color;

      el.innerHTML = `
        <svg viewBox="0 0 24 24" width="22" height="22">
          <path d="M12 2a2.5 2.5 0 0 0-2.5 2.5c0 1.6 1.2 2.5 2.5 2.5s2.5-.9 2.5-2.5A2.5 2.5 0 0 0 12 2z" fill="#1e1e24" stroke="#1e1e24" stroke-width="1"/>
          <path d="M17 12V10c0-1.2-1.2-2-2.8-2.4H9.8C8.2 8 7 8.8 7 10v2c0 .4.4.8.8.8h.8v6c0 .4.4.8.8.8h5.2c.4 0 .8-.4.8-.8v-6h.8c.4 0 .8-.4.8-.8z" fill="#1e1e24" stroke="#1e1e24" stroke-width="1"/>
        </svg>
      `;

      const host = meetup.host_name || meetup.hostName || "Ẩn danh";
      const count =
        meetup.players_count !== undefined
          ? meetup.players_count
          : meetup.playersCount || 1;
      const needed =
        meetup.players_needed !== undefined
          ? meetup.players_needed
          : meetup.playersNeeded || 4;

      const popupHTML = `
        <div class="cartoon-popup">
          <h4>${meetup.title}</h4>
          <p>🎲 Game: <strong>${meetup.game}</strong></p>
          <p>⏰ Lịch: <strong>${formatTime(meetup.time)}</strong></p>
          <p>👥 Sĩ số: <strong>${count}/${needed} người</strong></p>
          <p>👤 Chủ trì: <strong>${host}</strong></p>
          <button class="btn btn-join-meetup" onclick="alert('Đăng ký tham gia nhóm thành công!')">Tham gia hội</button>
        </div>
      `;

      const popup = new mapboxgl.Popup({
        offset: 25,
        closeButton: false,
      }).setHTML(popupHTML);

      // Stop click event propagation so meeple click opens popup without triggering map click
      el.addEventListener("click", (e) => {
        e.stopPropagation();
      });

      const marker = new mapboxgl.Marker(el)
        .setLngLat([meetup.lng, meetup.lat])
        .setPopup(popup)
        .addTo(map);

      markersList.push(marker);
      markersMap.set(meetup.id, marker);
    });
  }

  function formatTime(timeStr: string): string {
    if (!timeStr) return "";
    if (!timeStr.includes("T")) return timeStr;
    try {
      const date = new Date(timeStr);
      const day = String(date.getDate()).padStart(2, "0");
      const month = String(date.getMonth() + 1).padStart(2, "0");
      const year = date.getFullYear();
      const hours = String(date.getHours()).padStart(2, "0");
      const minutes = String(date.getMinutes()).padStart(2, "0");
      return `${hours}:${minutes}, Ngày ${day}/${month}/${year}`;
    } catch {
      return timeStr;
    }
  }

  // Reactive updates of markers based on filtered meetups prop (Svelte 5 effect)
  $effect(() => {
    if (mapInstance && (meetups || presetVenues)) {
      renderMarkers();
    }
  });

  // Reactive update of the temporary marker and camera flying when coordinates change from autocomplete input
  $effect(() => {
    if (mapInstance) {
      if (mode === "select" && selectedLat !== null && selectedLng !== null) {
        showTempMarker(selectedLng, selectedLat);
      } else {
        clearTempMarker();
      }
    }
  });

  // Automatically focus/fly to single meetup when viewing location of a specific meetup
  $effect(() => {
    if (mapInstance && mode === "discover" && meetups.length === 1) {
      mapInstance.flyTo({
        center: [meetups[0].lng, meetups[0].lat],
        zoom: 14.5,
        essential: true,
      });
    }
  });

  // Public methods to expose to parent
  export function clearTempMarker() {
    if (tempMarker) {
      tempMarker.remove();
      tempMarker = null;
    }
  }

  export function showTempMarker(lng: number, lat: number) {
    const map = mapInstance;
    if (!map) return;
    if (tempMarker) {
      tempMarker.remove();
    }

    const el = document.createElement("div");
    el.className = "custom-meeple-marker temp-marker-wrapper";
    el.style.backgroundColor = "#ffa4b2"; // pink red
    el.style.border = "2.5px solid #1e1e24"; // thick cartoon border
    el.style.boxShadow = "2px 2px 0 #1e1e24";
    el.style.width = "32px";
    el.style.height = "32px";
    el.style.borderRadius = "50%";
    el.style.display = "flex";
    el.style.alignItems = "center";
    el.style.justifyContent = "center";
    el.style.cursor = "pointer";

    // Wrap in a separate element to animate transform without breaking Mapbox positioning transform
    el.innerHTML = `
      <div class="temp-marker-pulse">
        <svg viewBox="0 0 24 24" width="22" height="22">
          <circle cx="12" cy="12" r="5" fill="#ef4444" />
        </svg>
      </div>
    `;

    tempMarker = new mapboxgl.Marker(el).setLngLat([lng, lat]).addTo(map);
  }

  export function flyToMeetup(meetup: Meetup) {
    if (!mapInstance) return;

    // Jump camera instantly on target meetup location without animation
    mapInstance.jumpTo({
      center: [meetup.lng, meetup.lat],
      zoom: 14.0,
    });

    // Toggle the popup associated with the marker
    const marker = markersMap.get(meetup.id);
    if (marker) {
      const popup = marker.getPopup();
      if (popup && !popup.isOpen()) {
        marker.togglePopup();
      }
      console.log(marker);
    }
  }

  async function initializeMap(token: string) {
    if (!token || !token.startsWith("pk.")) {
      isTokenMissing = true;
      return;
    }

    isTokenMissing = false;
    mapboxgl.accessToken = token;

    await tick();

    if (!mapContainer) {
      console.warn("Map container element not found yet. Retrying in 100ms...");
      setTimeout(() => initializeMap(token), 100);
      return;
    }

    if (mapInstance) {
      mapInstance.remove();
      mapInstance = null;
    }

    const initialCenter: [number, number] =
      selectedLng !== null && selectedLat !== null
        ? [selectedLng, selectedLat]
        : [106.692, 10.776];
    const initialZoom =
      selectedLng !== null && selectedLat !== null ? 13.5 : 12.2;

    try {
      mapInstance = new mapboxgl.Map({
        container: mapContainer,
        style: "mapbox://styles/mapbox/streets-v12",
        center: initialCenter,
        zoom: initialZoom,
      });

      // Add navigation controls
      mapInstance.addControl(new mapboxgl.NavigationControl(), "top-right");

      // Add geolocate control
      const geolocate = new mapboxgl.GeolocateControl({
        positionOptions: {
          enableHighAccuracy: true,
        },
        trackUserLocation: true,
        showUserHeading: true,
      });
      mapInstance.addControl(geolocate, "top-right");

      // Add click listener for selecting meetup location (only active in 'select' mode)
      mapInstance.on("click", (e) => {
        if (mode !== "select") return;
        const { lng, lat } = e.lngLat;

        // Show temp marker where user clicked geographically
        showTempMarker(lng, lat);

        // Notify parent callback
        if (onLocationSelect) {
          onLocationSelect(lng, lat);
        }
      });

      // Render the fetched meetups
      renderMarkers();
      isMapInitialized = true;
    } catch (err) {
      console.error("Mapbox GL initialization failed:", err);
    }
  }

  function handleManualInit(e: Event) {
    e.preventDefault();
    if (customToken.trim()) {
      initializeMap(customToken.trim());
    }
  }

  onMount(() => {
    const token = getMapboxToken();
    if (token) {
      initializeMap(token);
    } else {
      isTokenMissing = true;
    }
  });

  onDestroy(() => {
    if (mapInstance) {
      mapInstance.remove();
    }
  });
</script>

<div class="map-wrapper">
  {#if isTokenMissing}
    <!-- Playful Cartoon Warning Box when Token is Missing -->
    <div class="cartoon-card warning-map-card">
      <div class="warning-map-header">
        <span class="warning-map-icon">🗺️</span>
        <h3>Bản Đồ Cần Mapbox Access Token!</h3>
      </div>
      <p>
        Vui lòng thêm mã truy cập của bạn vào file <code>frontend/.env</code>:
      </p>
      <pre><code>VITE_MAPBOX_ACCESS_TOKEN=your_mapbox_token</code></pre>

      <div class="divider"><span>Hoặc nhập nhanh dưới đây để test</span></div>

      <form onsubmit={handleManualInit} class="manual-token-form">
        <input
          type="text"
          placeholder="Dán token pk.xxx của bạn vào đây..."
          bind:value={customToken}
          required
        />
        <button type="submit" class="btn btn-primary">Khởi tạo bản đồ 🚀</button
        >
      </form>
    </div>
  {:else}
    <!-- Mapbox container is rendered in DOM directly when token is available -->
    <div bind:this={mapContainer} class="map-container"></div>
  {/if}
</div>

<style>
  .map-wrapper {
    width: 100%;
    flex: 1;
    height: 100%;
    min-height: 220px;
    position: relative;
    border: var(--border-width) solid var(--color-border);
    border-radius: var(--radius-lg);
    overflow: hidden;
    background-color: var(--bg-secondary);
    box-shadow: var(--shadow-offset) var(--shadow-offset) 0 var(--color-border);
  }

  @media (max-width: 768px) {
    .map-wrapper {
      flex: 1;
      height: 100%;
      min-height: 200px;
    }
  }

  .map-container {
    width: 100%;
    height: 100%;
  }

  :global(.mapboxgl-canvas-container),
  :global(.mapboxgl-canvas) {
    width: 100% !important;
    height: 100% !important;
  }

  /* Warning container styles */
  .warning-map-card {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 90%;
    max-width: 480px;
    text-align: left;
    background-color: #fffefb;
    z-index: 10;
    box-shadow: 4px 4px 0 var(--color-border);
  }

  .warning-map-header {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
    color: var(--text-dark);
  }

  .warning-map-icon {
    font-size: 1.6rem;
  }

  .warning-map-card pre {
    background-color: var(--bg-secondary);
    border: 2px solid var(--color-border);
    padding: 8px 12px;
    border-radius: var(--radius-sm);
    font-family: monospace;
    font-size: 0.85rem;
    margin: 10px 0;
  }

  .manual-token-form {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .manual-token-form input {
    padding: 10px 14px;
    border-radius: var(--radius-md);
    border: var(--border-width) solid var(--color-border);
    font-family: var(--font-family);
    font-size: 0.95rem;
    font-weight: 600;
    box-shadow: 2px 2px 0 var(--color-border);
    outline: none;
  }

  .manual-token-form input:focus {
    transform: translate(-1px, -1px);
    box-shadow: 3px 3px 0 var(--color-border);
  }

  .divider {
    display: flex;
    align-items: center;
    text-align: center;
    margin: 15px 0;
    color: var(--text-muted);
    font-size: 0.85rem;
    font-weight: 700;
  }

  .divider::before,
  .divider::after {
    content: "";
    flex: 1;
    border-bottom: 1.5px dashed var(--color-border);
  }

  .divider span {
    padding: 0 8px;
  }

  /* Mapbox Custom Marker & Popup styling */
  :global(.custom-meeple-marker) {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    border: 2.5px solid #1e1e24;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    box-shadow: 2px 2px 0 #1e1e24;
    /* Limit transition properties to prevent lag in map coordinates positioning update */
    transition:
      background-color 0.15s ease-out,
      box-shadow 0.15s ease-out;
  }

  :global(.custom-meeple-marker:hover) {
    transform: scale(1.15) rotate(-5deg);
    box-shadow: 3px 3px 0 #1e1e24;
  }

  /* Temp marker child styling animation (isolates transform to prevent map layout break) */
  :global(.temp-marker-pulse) {
    display: flex;
    align-items: center;
    justify-content: center;
    animation: pulse-temp-marker 1.2s infinite ease-in-out;
  }

  @keyframes pulse-temp-marker {
    0% {
      transform: scale(1);
    }
    50% {
      transform: scale(1.25);
    }
    100% {
      transform: scale(1);
    }
  }

  /* Mapbox Popup override */
  :global(.mapboxgl-popup-content) {
    background: #fff !important;
    border: 2.5px solid #1e1e24 !important;
    border-radius: 16px !important;
    padding: 14px 16px !important;
    box-shadow: 4px 4px 0 #1e1e24 !important;
    font-family: "Quicksand", sans-serif !important;
  }

  :global(.mapboxgl-popup-tip) {
    border-top-color: #1e1e24 !important;
    border-bottom-color: #1e1e24 !important;
  }

  :global(.cartoon-popup h4) {
    font-weight: 700;
    font-size: 0.95rem;
    margin-bottom: 6px;
    color: #1e1e24;
  }

  :global(.cartoon-popup p) {
    font-size: 0.8rem;
    color: #646473;
    margin-bottom: 4px;
  }

  :global(.btn-join-meetup) {
    width: 100%;
    padding: 6px 12px;
    margin-top: 8px;
    font-size: 0.8rem;
    border-width: 2px;
    border-radius: 8px;
    background-color: #ffe869;
    font-weight: 700;
    cursor: pointer;
    border-style: solid;
    border-color: #1e1e24;
    box-shadow: 2px 2px 0 #1e1e24;
    transition: all 0.1s ease;
  }

  :global(.btn-join-meetup:hover) {
    transform: translate(-1px, -1px);
    box-shadow: 3px 3px 0 #1e1e24;
  }

  :global(.btn-join-meetup:active) {
    transform: translate(2px, 2px);
    box-shadow: 0px 0px 0 #1e1e24;
  }

  /* Cartoonish Mapbox controls (Zoom, Geolocate) */
  :global(.mapboxgl-ctrl-group) {
    border: 3px solid #1e1e24 !important;
    border-radius: 12px !important;
    box-shadow: 3px 3px 0 #1e1e24 !important;
    overflow: hidden;
    background-color: #fff !important;
    margin-top: 10px !important;
    margin-right: 10px !important;
  }
  :global(.mapboxgl-ctrl-group button) {
    border-bottom: 2px solid #1e1e24 !important;
    width: 32px !important;
    height: 32px !important;
  }
  :global(.mapboxgl-ctrl-group button:last-child) {
    border-bottom: none !important;
  }
</style>
