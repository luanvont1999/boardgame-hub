<script lang="ts">
  import { onMount, tick } from 'svelte';
  import Map from '../lib/Map.svelte';
  import { goBack, navigate } from '../lib/router.svelte';
  import Icon from '../lib/Icon.svelte';

  interface Props {
    meetups: any[];
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

  const popularVenues = [
    {
      name: "Boardgame Station",
      address: "21 Cô Bắc, Quận 1, TP. HCM",
      lat: 10.7656,
      lng: 106.6961,
    },
    {
      name: "Cashflow Cafe",
      address: "7A/19 Thành Thái, Quận 10, TP. HCM",
      lat: 10.7712,
      lng: 106.6644,
    },
    {
      name: "Ma Sói Guild",
      address: "12 Vệ Hồ, Tây Hồ, Hà Nội",
      lat: 21.062,
      lng: 105.8155,
    },
    {
      name: "The Boardgame Hub",
      address: "45 Lương Ngọc Quyến, Hoàn Kiếm, Hà Nội",
      lat: 21.0345,
      lng: 105.8524,
    },
  ];

  let mapComponent = $state<any>(null);
  let tempLat = $state<number | null>(null);
  let tempLng = $state<number | null>(null);
  let tempAddress = $state<string>('');
  let isResolvingAddress = $state<boolean>(false);

  onMount(() => {
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

  function selectPopularVenue(venue: typeof popularVenues[0]) {
    tempLat = venue.lat;
    tempLng = venue.lng;
    tempAddress = `${venue.name} (${venue.address})`;
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
      <h2>
        {#if mode === 'select'}
          <Icon name="gps" size={20} /> Chọn Vị Trí Lên Kèo
        {:else}
          <Icon name="map" size={20} /> Bản Đồ Hội Nhóm Boardgame
        {/if}
      </h2>
      <span class="sub-title">
        {mode === 'select'
          ? 'Nhấn chọn 1 điểm bất kỳ trên bản đồ hoặc chọn điểm gợi ý bên dưới'
          : 'Xem vị trí trực quan của tất cả các kèo chơi trên bản đồ'}
      </span>
    </div>
  </div>

  <!-- Popular Venue Shortcuts for Location Selection Mode -->
  {#if mode === 'select'}
    <div class="cartoon-card venue-shortcuts-bar">
      <span class="venue-shortcuts-title">
        <Icon name="sparkles" size={15} /> Gợi ý nhanh điểm chơi:
      </span>
      <div class="venue-tags">
        {#each popularVenues as venue}
          <button
            type="button"
            class="venue-tag-btn {tempLat === venue.lat && tempLng === venue.lng ? 'active' : ''}"
            onclick={() => selectPopularVenue(venue)}
          >
            <Icon name="store" size={13} /> {venue.name}
          </button>
        {/each}
      </div>
    </div>
  {/if}

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
          <span><Icon name="pin" size={16} /> Địa chỉ đã chọn: <strong>{isResolvingAddress ? 'Đang xác định...' : tempAddress}</strong></span>
        {:else}
          <span>Vui lòng chọn 1 điểm gợi ý trên hoặc chạm vào bản đồ.</span>
        {/if}
      </div>
      <button
        type="button"
        class="btn btn-primary btn-confirm"
        disabled={tempLat === null || isResolvingAddress}
        onclick={confirmLocation}
      >
        <Icon name="check-circle" size={16} /> Xác nhận vị trí này
      </button>
    {:else}
      <span class="map-hint-text">
        <Icon name="sparkles" size={16} /> Bấm vào quân cờ Meeple màu sắc trên bản đồ để xem chi tiết kèo.
      </span>
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
    gap: 12px;
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
    padding: 12px 20px;
    border-radius: var(--radius-lg);
    box-shadow: 4px 4px 0 var(--color-border);
    text-align: left;
    flex-shrink: 0;
  }

  .back-btn { padding: 8px 16px !important; font-size: 0.95rem !important; white-space: nowrap; }

  .nav-title-group h2 {
    font-size: 1.2rem;
    font-weight: 800;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .sub-title { font-size: 0.82rem; font-weight: 600; color: var(--text-dark); }

  /* Venue Shortcuts Bar */
  .venue-shortcuts-bar {
    background-color: #fffefb;
    padding: 10px 16px;
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
    flex-shrink: 0;
    text-align: left;
  }

  .venue-shortcuts-title {
    font-size: 0.85rem;
    font-weight: 800;
    color: var(--text-dark);
    display: inline-flex;
    align-items: center;
    gap: 4px;
    white-space: nowrap;
  }

  .venue-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .venue-tag-btn {
    font-size: 0.8rem;
    font-weight: 700;
    font-family: var(--font-family);
    padding: 6px 12px;
    border: 2px solid #1e1e24;
    border-radius: 100px;
    background-color: #ffffff;
    box-shadow: 2px 2px 0px #1e1e24;
    cursor: pointer;
    transition: all 0.1s ease;
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .venue-tag-btn:hover {
    transform: translate(-1px, -1px);
    box-shadow: 3px 3px 0px #1e1e24;
    background-color: var(--pastel-yellow, #ffe869);
  }

  .venue-tag-btn.active {
    background-color: var(--pastel-yellow, #ffe869);
    box-shadow: 1px 1px 0px #1e1e24;
    transform: translate(1px, 1px);
  }

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
    padding: 12px 20px;
    border-radius: var(--radius-lg);
    box-shadow: 4px 4px 0 var(--color-border);
    flex-shrink: 0;
    gap: 16px;
  }

  .location-picker-info {
    text-align: left;
    font-size: 0.88rem;
    font-weight: 700;
    color: var(--text-dark);
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .map-hint-text {
    font-size: 0.88rem;
    font-weight: 700;
    color: var(--text-muted);
    text-align: left;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .btn-confirm {
    padding: 10px 20px !important;
    font-size: 0.95rem !important;
    white-space: nowrap;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
</style>
