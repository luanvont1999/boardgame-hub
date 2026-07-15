<script lang="ts">
  import { onMount, tick } from "svelte";
  import Map from "../lib/Map.svelte";
  import { goBack, navigate } from "../lib/router.svelte";
  import Icon from "../lib/Icon.svelte";

  interface Props {
    meetups: any[];
    selectedLat: number | null;
    selectedLng: number | null;
    addressText: string;
    mode: "discover" | "select";
    meetupId?: string;
    userLat?: number | null;
    userLng?: number | null;
  }

  let {
    meetups,
    selectedLat = $bindable(null),
    selectedLng = $bindable(null),
    addressText = $bindable(""),
    mode,
    meetupId,
    userLat = null,
    userLng = null,
  }: Props = $props();

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

  let singleMeetup = $derived(
    meetupId ? meetups.find((m) => m.id === meetupId) : null,
  );

  let displayedMeetups = $derived.by(() => {
    if (mode === "select") return [];
    if (meetupId) return singleMeetup ? [singleMeetup] : [];
    return meetups;
  });

  const PREDEFINED_VENUES = [
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
      name: "The Guild Board Game",
      address: "188/1 Nguyễn Văn Hưởng, Thảo Điền, TP. Thủ Đức",
      lat: 10.8062,
      lng: 106.7325,
    },
    {
      name: "Meeple Den Cafe",
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
    {
      name: "The Nest Boardgame",
      address: "4A Tràng Thi, Hoàn Kiếm, Hà Nội",
      lat: 21.0288,
      lng: 105.8475,
    },
  ];

  let sortedVenues = $derived.by(() => {
    if (userLat !== null && userLng !== null) {
      const withDistance = PREDEFINED_VENUES.map((v) => ({
        ...v,
        distance: calculateDistance(userLat, userLng, v.lat, v.lng),
      }));

      // Filter strictly venues < 10km away from user
      const under10km = withDistance
        .filter((v) => v.distance < 10)
        .sort((a, b) => a.distance - b.distance);

      if (under10km.length > 0) {
        return under10km;
      }

      // Fallback: If no venue is under 10km, return the top 3 closest venues
      return withDistance.sort((a, b) => a.distance - b.distance).slice(0, 3);
    }
    return PREDEFINED_VENUES.map((v) => ({ ...v, distance: null }));
  });

  let mapComponent = $state<any>(null);
  let tempLat = $state<number | null>(null);
  let tempLng = $state<number | null>(null);
  let tempAddress = $state<string>("");
  let isResolvingAddress = $state<boolean>(false);

  onMount(() => {
    if (mode === "select" && selectedLat !== null) {
      tempLat = selectedLat;
      tempLng = selectedLng;
      tempAddress = addressText;
    }
  });

  function handleMapClick(lng: number, lat: number) {
    tempLat = lat;
    tempLng = lng;
    if (mode === "select") {
      reverseGeocode(lat, lng);
    }
  }

  function selectPopularVenue(venue: (typeof sortedVenues)[0]) {
    tempLat = venue.lat;
    tempLng = venue.lng;
    tempAddress = `${venue.name} (${venue.address})`;
  }

  async function reverseGeocode(lat: number, lng: number) {
    const token = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;
    if (!token) return;
    isResolvingAddress = true;
    tempAddress = "Đang xác định địa chỉ...";
    try {
      const res = await fetch(
        `https://api.mapbox.com/geocoding/v5/mapbox.places/${lng},${lat}.json?access_token=${token}&country=vn&limit=1&language=vi`,
      );
      if (res.ok) {
        const data = await res.json();
        tempAddress =
          data.features?.[0]?.place_name ??
          `${lat.toFixed(5)}, ${lng.toFixed(5)}`;
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
        {#if mode === "select"}
          <Icon name="gps" size={20} /> Chọn Vị Trí Lên Kèo
        {:else if singleMeetup}
          <Icon name="map" size={20} /> Vị Trí Kèo: {singleMeetup.title}
        {:else}
          <Icon name="map" size={20} /> Bản Đồ Hội Nhóm Boardgame
        {/if}
      </h2>
      <span class="sub-title">
        {#if mode === "select"}
          Nhấn chọn 1 điểm bất kỳ trên bản đồ hoặc chọn điểm gợi ý bên dưới
        {:else if singleMeetup}
          Game: <strong>{singleMeetup.game}</strong> • Host:
          <strong
            >{singleMeetup.hostName ||
              singleMeetup.host_name ||
              "Ẩn danh"}</strong
          >
        {:else}
          Xem vị trí trực quan của tất cả các kèo chơi trên bản đồ
        {/if}
      </span>
    </div>
  </div>

  <!-- Popular Venue Shortcuts for Location Selection Mode -->
  {#if mode === "select"}
    <div class="cartoon-card venue-shortcuts-bar">
      <p class="venue-shortcuts-title">
        <Icon name="sparkles" size={15} /> Gợi ý điểm chơi (&lt; 10km):
      </p>
      <div class="venue-tags">
        {#each sortedVenues as venue}
          <button
            type="button"
            class="venue-tag-btn {tempLat === venue.lat && tempLng === venue.lng
              ? 'active'
              : ''}"
            onclick={() => selectPopularVenue(venue)}
          >
            <Icon name="store" size={13} />
            <span>{venue.name}</span>
            {#if venue.distance !== null}
              <span class="venue-dist-tag"
                >• {venue.distance < 1
                  ? `${Math.round(venue.distance * 1000)}m`
                  : `${venue.distance.toFixed(1)}km`}</span
              >
            {/if}
          </button>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Interactive Map Canvas -->
  <Map
    bind:this={mapComponent}
    meetups={displayedMeetups}
    presetVenues={mode === "select" ? sortedVenues : []}
    selectedLat={tempLat}
    selectedLng={tempLng}
    {mode}
    onLocationSelect={handleMapClick}
  />

  <!-- Bottom Control Bar -->
  <div class="cartoon-card map-bottom-bar">
    {#if mode === "select"}
      <div class="location-picker-info">
        {#if tempLat && tempLng}
          <span
            ><Icon name="pin" size={16} /> Địa chỉ đã chọn:
            <strong
              >{isResolvingAddress ? "Đang xác định..." : tempAddress}</strong
            ></span
          >
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
        <Icon name="sparkles" size={16} /> Bấm vào quân cờ Meeple màu sắc trên bản
        đồ để xem chi tiết kèo.
      </span>
      <button type="button" class="btn btn-primary" onclick={goBack}
        >Xong / Trở về</button
      >
    {/if}
  </div>
</div>

<style>
  .fullscreen-route-view {
    width: 100%;
    display: flex;
    flex-direction: column;
    height: calc(100vh - 90px);
    max-height: calc(100vh - 90px);
    gap: 10px;
    margin-bottom: 0;
  }

  @media (max-width: 768px) {
    .fullscreen-route-view {
      height: calc(100vh - 90px);
      max-height: calc(100vh - 90px);
    }
  }

  .route-top-nav {
    background-color: var(--pastel-cyan);
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 14px;
    border-radius: var(--radius-md);
    box-shadow: 3px 3px 0 var(--color-border);
    text-align: left;
    flex-shrink: 0;
  }

  .back-btn {
    padding: 6px 12px !important;
    font-size: 0.85rem !important;
    white-space: nowrap;
  }

  .nav-title-group h2 {
    font-size: 1.05rem;
    font-weight: 800;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .sub-title {
    font-size: 0.78rem;
    font-weight: 600;
    color: var(--text-dark);
    display: block;
    line-height: 1.2;
  }

  /* Venue Shortcuts Bar with Horizontal Scroll */
  .venue-shortcuts-bar {
    background-color: #fffefb;
    padding: 8px 12px;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
    text-align: left;
    overflow: hidden;
  }

  .venue-shortcuts-title {
    font-size: 0.8rem;
    font-weight: 800;
    color: var(--text-dark);
    display: inline-flex;
    align-items: center;
    gap: 4px;
    white-space: nowrap;
    flex-shrink: 0;
  }

  .venue-tags {
    display: flex;
    flex-wrap: nowrap;
    overflow-x: auto;
    scrollbar-width: none;
    -webkit-overflow-scrolling: touch;
    gap: 6px;
    padding: 2px 0;
    width: 100%;
  }

  .venue-tags::-webkit-scrollbar {
    display: none;
  }

  .venue-tag-btn {
    font-size: 0.75rem;
    font-weight: 700;
    font-family: var(--font-family);
    padding: 4px 10px;
    border: 1.5px solid #1e1e24;
    border-radius: 100px;
    background-color: #ffffff;
    box-shadow: 1.5px 1.5px 0px #1e1e24;
    cursor: pointer;
    transition: all 0.1s ease;
    display: inline-flex;
    align-items: center;
    gap: 4px;
    flex-shrink: 0;
    white-space: nowrap;
  }

  .venue-tag-btn:hover {
    transform: translate(-1px, -1px);
    box-shadow: 2px 2px 0px #1e1e24;
    background-color: var(--pastel-yellow, #ffe869);
  }

  .venue-tag-btn.active {
    background-color: var(--pastel-yellow, #ffe869);
    box-shadow: 1px 1px 0px #1e1e24;
    transform: translate(1px, 1px);
  }

  .venue-dist-tag {
    font-size: 0.7rem;
    opacity: 0.85;
  }

  .map-bottom-bar {
    background-color: #fff;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 14px;
    border-radius: var(--radius-md);
    box-shadow: 3px 3px 0 var(--color-border);
    flex-shrink: 0;
    gap: 12px;
  }

  .location-picker-info {
    text-align: left;
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--text-dark);
    flex: 1;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .map-hint-text {
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--text-muted);
    text-align: left;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .btn-confirm {
    padding: 8px 16px !important;
    font-size: 0.88rem !important;
    white-space: nowrap;
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
</style>
