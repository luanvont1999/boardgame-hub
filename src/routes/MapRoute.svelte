<script lang="ts">
  import { onMount, tick } from 'svelte';
  import Map from '../lib/Map.svelte';
  import { goBack, navigate } from '../lib/router';

  interface Props {
    meetups: any[];
    // For 'select' mode — bindable so CreateRoute can read the picked coords
    selectedLat: number | null;
    selectedLng: number | null;
    addressText: string;
    mode: 'discover' | 'select';
  }

  let { 
    meetups,
    selectedLat = $bindable(null),
    selectedLng = $bindable(null),
    addressText = $bindable(''),
    mode
  }: Props = $props();

  let mapComponent = $state<any>(null);
  let tempLat = $state<number | null>(null);
  let tempLng = $state<number | null>(null);
  let tempAddress = $state<string>('');
  let isResolvingAddress = $state<boolean>(false);

  onMount(() => {
    // Seed temp from existing selection when in select mode
    if (mode === 'select' && selectedLat !== null) {
      tempLat = selectedLat;
      tempLng = selectedLng;
      tempAddress = addressText;
    }
  });

  function handleMapClick(lng: number, lat: number) {
    tempLat = lat;
    tempLng = lng;
    if (mode === 'select') {
      reverseGeocode(lat, lng);
    }
  }

  async function reverseGeocode(lat: number, lng: number) {
    const token = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;
    if (!token) return;
    isResolvingAddress = true;
    tempAddress = 'Đang xác định địa chỉ...';
    try {
      const res = await fetch(
        `https://api.mapbox.com/geocoding/v5/mapbox.places/${lng},${lat}.json?access_token=${token}&country=vn&limit=1&language=vi`
      );
      if (res.ok) {
        const data = await res.json();
        tempAddress = data.features?.[0]?.place_name ?? `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
      }
    } catch {
      tempAddress = `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
    } finally {
      isResolvingAddress = false;
    }
  }

  function confirmLocation() {
    // Write back to bindable props, then navigate back
    selectedLat = tempLat;
    selectedLng = tempLng;
    addressText = tempAddress;
    goBack();
  }
</script>

<div class="fullscreen-route-view map-route">
  <!-- Top Navigation -->
  <div class="cartoon-card route-top-nav">
    <button type="button" class="btn btn-secondary back-btn" onclick={goBack}>
      ← Quay lại
    </button>
    <div class="nav-title-group">
      <h2>{mode === 'select' ? '🎯 Chọn Vị Trí Lên Kèo' : '🗺️ Bản Đồ Hội Nhóm Boardgame'}</h2>
      <span class="sub-title">
        {mode === 'select'
          ? 'Nhấn chọn 1 điểm bất kỳ trên bản đồ để xác định địa điểm chơi'
          : 'Xem vị trí trực quan của tất cả các kèo chơi trên bản đồ'}
      </span>
    </div>
  </div>

  <!-- Interactive Map Canvas -->
  <div class="cartoon-card map-canvas-card">
    <Map
      bind:this={mapComponent}
      {meetups}
      selectedLat={tempLat}
      selectedLng={tempLng}
      onLocationSelect={handleMapClick}
    />
  </div>

  <!-- Bottom Control Bar -->
  <div class="cartoon-card map-bottom-bar">
    {#if mode === 'select'}
      <div class="location-picker-info">
        {#if tempLat && tempLng}
          <span>📍 Địa chỉ đã chọn: <strong>{isResolvingAddress ? 'Đang xác định...' : tempAddress}</strong></span>
        {:else}
          <span>👉 Vui lòng chạm hoặc click lên điểm bất kỳ trên bản đồ.</span>
        {/if}
      </div>
      <button
        type="button"
        class="btn btn-primary btn-confirm"
        disabled={tempLat === null || isResolvingAddress}
        onclick={confirmLocation}
      >
        Xác nhận vị trí này 🎯
      </button>
    {:else}
      <span class="map-hint-text">💡 Bấm vào quân cờ Meeple màu sắc trên bản đồ để xem chi tiết kèo.</span>
      <button type="button" class="btn btn-primary" onclick={goBack}>Xong / Trở về</button>
    {/if}
  </div>
</div>

<style>
  .fullscreen-route-view {
    width: 100%;
    display: flex;
    flex-direction: column;
    height: calc(100vh - 80px);
    gap: 16px;
    margin-bottom: 20px;
  }
  @media (max-width: 768px) {
    .fullscreen-route-view { height: calc(100vh - 80px); }
  }
  .route-top-nav {
    background-color: var(--pastel-cyan);
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 14px 20px;
    border-radius: var(--radius-lg);
    box-shadow: 4px 4px 0 var(--color-border);
    text-align: left;
    flex-shrink: 0;
  }
  .back-btn { padding: 8px 16px !important; font-size: 0.95rem !important; white-space: nowrap; }
  .nav-title-group h2 { font-size: 1.25rem; font-weight: 800; margin: 0; }
  .sub-title { font-size: 0.82rem; font-weight: 600; color: var(--text-dark); }
  .map-canvas-card {
    flex: 1;
    padding: 0 !important;
    overflow: hidden;
    position: relative;
    border-radius: var(--radius-lg);
    box-shadow: 4px 4px 0 var(--color-border);
  }
  .map-bottom-bar {
    background-color: #fff;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 20px;
    border-radius: var(--radius-lg);
    box-shadow: 4px 4px 0 var(--color-border);
    flex-shrink: 0;
    gap: 16px;
  }
  .location-picker-info {
    text-align: left;
    font-size: 0.9rem;
    font-weight: 700;
    color: var(--text-dark);
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .map-hint-text { font-size: 0.88rem; font-weight: 700; color: var(--text-muted); text-align: left; }
  .btn-confirm { padding: 10px 20px !important; font-size: 0.95rem !important; white-space: nowrap; }
</style>
