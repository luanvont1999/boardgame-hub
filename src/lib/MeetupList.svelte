<script lang="ts">
  import { onMount } from 'svelte';

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

  // Svelte 5 Bindable and Normal Props
  interface Props {
    meetups: Meetup[];
    userLat: number | null;
    userLng: number | null;
    selectedCity: 'all' | 'HCM' | 'HN';
    selectedDistance: 'all' | '5' | '10';
    isTrackingGPS: boolean;
    gpsError: boolean;
    onSelectMeetup: (meetup: Meetup) => void;
  }

  let { 
    meetups, 
    userLat, 
    userLng, 
    selectedCity = $bindable('all'), 
    selectedDistance = $bindable('all'),
    isTrackingGPS,
    gpsError,
    onSelectMeetup
  }: Props = $props();

  function calculateDistance(lat1: number, lon1: number, lat2: number, lon2: number): number {
    const R = 6371; // km
    const dLat = (lat2 - lat1) * Math.PI / 180;
    const dLon = (lon2 - lon1) * Math.PI / 180;
    const a = 
      Math.sin(dLat/2) * Math.sin(dLat/2) +
      Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) * 
      Math.sin(dLon/2) * Math.sin(dLon/2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1-a));
    return R * c;
  }

  function getMeetupCity(meetup: Meetup): 'HCM' | 'HN' {
    const distToHCM = calculateDistance(meetup.lat, meetup.lng, 10.7769, 106.7009);
    const distToHN = calculateDistance(meetup.lat, meetup.lng, 21.0285, 105.8542);
    return distToHN < distToHCM ? 'HN' : 'HCM';
  }

  function formatTime(timeStr: string): string {
    if (!timeStr) return '';
    if (!timeStr.includes('T')) return timeStr;
    try {
      const date = new Date(timeStr);
      const day = String(date.getDate()).padStart(2, '0');
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const year = date.getFullYear();
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      return `${hours}:${minutes}, ${day}/${month}/${year}`;
    } catch {
      return timeStr;
    }
  }
</script>

<div class="meetup-list-container">
  <!-- Cartoon Control Filter Bar -->
  <div class="cartoon-card filter-bar-card">
    <div class="filter-header">
      <span class="filter-title-icon">🔍</span>
      <h3>Bộ lọc tìm kiếm kèo chơi</h3>
    </div>

    <div class="filter-groups">
      <!-- City Filter Group -->
      <div class="filter-group">
        <span class="filter-label">Khu vực:</span>
        <div class="filter-buttons">
          <button 
            class="filter-btn {selectedCity === 'all' ? 'active' : ''}" 
            onclick={() => selectedCity = 'all'}
          >
            Tất cả
          </button>
          <button 
            class="filter-btn {selectedCity === 'HCM' ? 'active' : ''}" 
            onclick={() => selectedCity = 'HCM'}
          >
            TP. HCM
          </button>
          <button 
            class="filter-btn {selectedCity === 'HN' ? 'active' : ''}" 
            onclick={() => selectedCity = 'HN'}
          >
            Hà Nội
          </button>
        </div>
      </div>

      <!-- Distance Filter Group -->
      <div class="filter-group">
        <span class="filter-label">Khoảng cách:</span>
        <div class="filter-buttons">
          <button 
            class="filter-btn {selectedDistance === 'all' ? 'active' : ''}" 
            onclick={() => selectedDistance = 'all'}
          >
            Mọi khoảng cách
          </button>
          <button 
            class="filter-btn {selectedDistance === '5' ? 'active' : ''}" 
            disabled={userLat === null}
            onclick={() => selectedDistance = '5'}
          >
            &lt; 5 km
          </button>
          <button 
            class="filter-btn {selectedDistance === '10' ? 'active' : ''}" 
            disabled={userLat === null}
            onclick={() => selectedDistance = '10'}
          >
            &lt; 10 km
          </button>
        </div>
      </div>
    </div>

    <!-- Geolocation Status badge -->
    <div class="gps-status-bar">
      {#if isTrackingGPS}
        <span class="gps-dot pulsing"></span>
        <span class="gps-text">Đang lấy vị trí GPS...</span>
      {:else if userLat !== null}
        <span class="gps-dot success"></span>
        <span class="gps-text">Đã bật định vị GPS (Khoảng cách hoạt động)</span>
      {:else if gpsError}
        <span class="gps-dot warning"></span>
        <span class="gps-text">GPS bị từ chối hoặc không hỗ trợ. (Vui lòng bật GPS trên Mapbox để lọc theo khoảng cách)</span>
      {/if}
    </div>
  </div>

  <!-- Meetups Cards Grid -->
  <div class="meetup-cards-grid">
    {#if meetups.length === 0}
      <div class="cartoon-card no-meetup-card">
        <h3>Không tìm thấy kèo chơi nào phù hợp! 😢</h3>
        <p>Thử đổi bộ lọc hoặc nhấp chuột lên bản đồ để tự lên kèo mới nhé!</p>
      </div>
    {:else}
      {#each meetups as meetup (meetup.id)}
        {@const city = getMeetupCity(meetup)}
        {@const distance = userLat !== null && userLng !== null ? calculateDistance(userLat, userLng, meetup.lat, meetup.lng) : null}
        
        <div class="cartoon-card meetup-item-card" style="border-top: 10px solid {meetup.color};">
          <!-- Card Badge info -->
          <div class="meetup-card-badges">
            <span class="city-badge {city === 'HCM' ? 'hcm-color' : 'hn-color'}">
              {city === 'HCM' ? '🌆 HCM' : '🏰 HN'}
            </span>
            {#if distance !== null}
              <span class="distance-badge">
                📍 Cách {distance.toFixed(1)} km
              </span>
            {/if}
          </div>

          <h3 class="meetup-card-title">{meetup.title}</h3>
          
          <div class="meetup-card-details">
            <div class="detail-row">
              <span class="detail-icon">🎲</span>
              <span>Game: <strong>{meetup.game}</strong></span>
            </div>
            <div class="detail-row">
              <span class="detail-icon">⏰</span>
              <span>Lịch: <strong>{formatTime(meetup.time)}</strong></span>
            </div>
            <div class="detail-row">
              <span class="detail-icon">👥</span>
              <span>Sĩ số: <strong>{meetup.players_count || meetup.playersCount || 1} / {meetup.players_needed || meetup.playersNeeded || 4} người</strong></span>
            </div>
            <div class="detail-row">
              <span class="detail-icon">👤</span>
              <span>Host: <strong>{meetup.host_name || meetup.hostName || 'Ẩn danh'}</strong></span>
            </div>
          </div>

          <button class="btn btn-secondary w-full" onclick={() => onSelectMeetup(meetup)}>
            Xem vị trí trên bản đồ 🗺️
          </button>
        </div>
      {/each}
    {/if}
  </div>
</div>

<style>
  .meetup-list-container {
    width: 100%;
    margin-top: 24px;
    display: flex;
    flex-direction: column;
    gap: 32px;
  }

  .filter-bar-card {
    background-color: #fffefb;
    text-align: left;
    padding: 20px;
  }

  .filter-header {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;
    color: var(--text-dark);
  }

  .filter-title-icon {
    font-size: 1.4rem;
  }

  .filter-groups {
    display: flex;
    flex-wrap: wrap;
    gap: 24px;
  }

  .filter-group {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-wrap: wrap;
  }

  .filter-label {
    font-weight: 700;
    font-size: 0.95rem;
    color: var(--text-dark);
  }

  .filter-buttons {
    display: flex;
    gap: 8px;
  }

  .filter-btn {
    padding: 6px 14px;
    font-size: 0.9rem;
    font-weight: 700;
    border-radius: var(--radius-sm);
    border: var(--border-width-sm) solid var(--color-border);
    background-color: #fff;
    cursor: pointer;
    box-shadow: 2px 2px 0 var(--color-border);
    font-family: var(--font-family);
    transition: all 0.1s ease;
  }

  .filter-btn:hover:not(:disabled) {
    transform: translate(-1px, -1px);
    box-shadow: 3px 3px 0 var(--color-border);
  }

  .filter-btn:active:not(:disabled) {
    transform: translate(2px, 2px);
    box-shadow: 0 0 0 var(--color-border);
  }

  .filter-btn.active {
    background-color: var(--pastel-yellow);
  }

  .filter-btn:disabled {
    background-color: #e5e7eb;
    color: #9ca3af;
    border-color: #d1d5db;
    cursor: not-allowed;
    box-shadow: 1px 1px 0 #d1d5db;
  }

  /* GPS status bar */
  .gps-status-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 16px;
    padding-top: 12px;
    border-top: 1.5px dashed var(--color-border);
  }

  .gps-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    border: 1px solid var(--color-border);
    display: inline-block;
  }

  .gps-dot.pulsing {
    background-color: #eab308;
    animation: gps-pulse 1s infinite;
  }

  .gps-dot.success {
    background-color: #10b981;
  }

  .gps-dot.warning {
    background-color: #ef4444;
  }

  @keyframes gps-pulse {
    0% { transform: scale(0.9); opacity: 0.6; }
    50% { transform: scale(1.15); opacity: 1; }
    100% { transform: scale(0.9); opacity: 0.6; }
  }

  .gps-text {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-muted);
  }

  /* Cards Grid */
  .meetup-cards-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 24px;
  }

  @media (max-width: 480px) {
    .meetup-cards-grid {
      grid-template-columns: 1fr;
    }
  }

  .no-meetup-card {
    grid-column: 1 / -1;
    text-align: center;
    background-color: #fff;
    padding: 40px 20px;
  }

  .no-meetup-card p {
    color: var(--text-muted);
    font-weight: 500;
    margin-top: 8px;
  }

  /* Item Card styling */
  .meetup-item-card {
    background-color: #fffefc;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    text-align: left;
    padding: 20px;
  }

  .meetup-card-badges {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
  }

  .city-badge, .distance-badge {
    font-size: 0.8rem;
    font-weight: 700;
    padding: 4px 10px;
    border-radius: 100px;
    border: var(--border-width-sm) solid var(--color-border);
    box-shadow: 1.5px 1.5px 0 var(--color-border);
  }

  .city-badge.hcm-color {
    background-color: var(--pastel-pink);
  }

  .city-badge.hn-color {
    background-color: var(--pastel-green);
  }

  .distance-badge {
    background-color: var(--pastel-cyan);
  }

  .meetup-card-title {
    font-size: 1.25rem;
    font-weight: 700;
    margin-bottom: 16px;
    line-height: 1.3;
    color: var(--text-dark);
  }

  .meetup-card-details {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 20px;
  }

  .detail-row {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.9rem;
    color: var(--text-dark);
    font-weight: 500;
  }

  .detail-icon {
    font-size: 1.1rem;
  }

  .w-full {
    width: 100%;
  }
</style>
