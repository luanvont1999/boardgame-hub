<script lang="ts">
  import { goBack } from '../lib/router.svelte';
  import Icon from '../lib/Icon.svelte';

  interface Props {
    selectedCity: 'all' | 'HCM' | 'HN';
    selectedDistance: 'all' | '5' | '10';
    userLat: number | null;
    isTrackingGPS: boolean;
    gpsError: boolean;
    onApply: (city: 'all' | 'HCM' | 'HN', distance: 'all' | '5' | '10') => void;
  }

  let { 
    selectedCity = 'all', 
    selectedDistance = 'all', 
    userLat, 
    isTrackingGPS, 
    gpsError,
    onApply,
  }: Props = $props();

  let tempCity = $state<'all' | 'HCM' | 'HN'>('all');
  let tempDistance = $state<'all' | '5' | '10'>('all');

  $effect(() => {
    tempCity = selectedCity;
    tempDistance = selectedDistance;
  });

  function handleApply() {
    onApply(tempCity, tempDistance);
    goBack();
  }
</script>

<div class="fullscreen-route-view filter-route">
  <!-- Top Navigation -->
  <div class="cartoon-card route-top-nav">
    <button type="button" class="btn btn-secondary back-btn" onclick={goBack}>
      ← Quay lại
    </button>
    <div class="nav-title-group">
      <h2><Icon name="filter" size={20} /> Tùy Chỉnh Bộ Lọc Tìm Kiếm</h2>
      <span class="sub-title">Chọn khu vực thành phố và bán kính vị trí</span>
    </div>
  </div>

  <!-- Compact Body Content -->
  <div class="cartoon-card filter-content-card">
    <!-- City Selection Row -->
    <div class="filter-section">
      <h3 class="section-title-label">
        <Icon name="building" size={16} /> Khu Vực:
      </h3>
      <div class="options-row">
        <button 
          type="button" 
          class="option-chip-btn {tempCity === 'all' ? 'active' : ''}"
          onclick={() => tempCity = 'all'}
        >
          <Icon name="map" size={15} />
          <span>Tất Cả</span>
        </button>

        <button 
          type="button" 
          class="option-chip-btn {tempCity === 'HCM' ? 'active' : ''}"
          onclick={() => tempCity = 'HCM'}
        >
          <Icon name="building" size={15} />
          <span>Hồ Chí Minh</span>
        </button>

        <button 
          type="button" 
          class="option-chip-btn {tempCity === 'HN' ? 'active' : ''}"
          onclick={() => tempCity = 'HN'}
        >
          <Icon name="landmark" size={15} />
          <span>Hà Nội</span>
        </button>
      </div>
    </div>

    <!-- Distance Selection Row -->
    <div class="filter-section" style="margin-top: 20px;">
      <h3 class="section-title-label">
        <Icon name="gps" size={16} /> Bán Kính (GPS):
      </h3>
      <div class="options-row">
        <button 
          type="button" 
          class="option-chip-btn {tempDistance === 'all' ? 'active' : ''}"
          onclick={() => tempDistance = 'all'}
        >
          <Icon name="pin" size={15} />
          <span>Tất Cả</span>
        </button>

        <button 
          type="button" 
          class="option-chip-btn {tempDistance === '5' ? 'active' : ''}"
          disabled={userLat === null}
          onclick={() => tempDistance = '5'}
        >
          <Icon name="zap" size={15} />
          <span>&lt; 5 km</span>
        </button>

        <button 
          type="button" 
          class="option-chip-btn {tempDistance === '10' ? 'active' : ''}"
          disabled={userLat === null}
          onclick={() => tempDistance = '10'}
        >
          <Icon name="rocket" size={15} />
          <span>&lt; 10 km</span>
        </button>
      </div>

      {#if userLat === null}
        <p class="gps-warning-hint">
          <Icon name="alert-triangle" size={14} /> Cần bật vị trí (GPS) để xem bán kính theo km.
        </p>
      {/if}
    </div>

    <!-- GPS Status bar -->
    <div class="gps-status-pill">
      {#if isTrackingGPS}
        <span class="gps-dot pulsing"></span>
        <span class="gps-text">Đang lấy vị trí GPS...</span>
      {:else if userLat !== null}
        <span class="gps-dot success"></span>
        <span class="gps-text">Đã định vị thành công</span>
      {:else if gpsError}
        <span class="gps-dot warning"></span>
        <span class="gps-text">Chưa bật định vị GPS</span>
      {/if}
    </div>

    <div class="apply-action-bar">
      <button type="button" class="btn btn-primary btn-apply-full" onclick={handleApply}>
        <Icon name="check-circle" size={18} /> Áp Dụng Bộ Lọc
      </button>
    </div>
  </div>
</div>

<style>
  .fullscreen-route-view {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding-bottom: 40px;
    max-width: 580px;
    margin: 0 auto;
  }

  .route-top-nav {
    background-color: var(--pastel-yellow);
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    border-radius: var(--radius-lg);
    box-shadow: 3px 3px 0 var(--color-border);
    text-align: left;
  }

  .back-btn {
    padding: 6px 12px !important;
    font-size: 0.88rem !important;
    white-space: nowrap;
  }

  .nav-title-group h2 {
    font-size: 1.1rem;
    font-weight: 800;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .sub-title {
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text-dark);
  }

  .filter-content-card {
    background-color: #fffefb;
    padding: 24px 20px;
    text-align: left;
  }

  .section-title-label {
    font-size: 0.95rem;
    font-weight: 800;
    margin-bottom: 10px;
    color: var(--text-dark);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  /* Compact Horizontal Options Row */
  .options-row {
    display: flex;
    gap: 8px;
    width: 100%;
  }

  .option-chip-btn {
    flex: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 10px 8px;
    border-radius: var(--radius-md);
    border: var(--border-width) solid var(--color-border);
    background-color: #fff;
    cursor: pointer;
    box-shadow: 2.5px 2.5px 0 var(--color-border);
    font-family: var(--font-family);
    font-size: 0.88rem;
    font-weight: 800;
    white-space: nowrap;
    transition: all 0.1s ease;
  }

  .option-chip-btn:hover:not(:disabled) {
    transform: translate(-1px, -1px);
    box-shadow: 3.5px 3.5px 0 var(--color-border);
  }

  .option-chip-btn:active:not(:disabled) {
    transform: translate(1px, 1px);
    box-shadow: 1px 1px 0 var(--color-border);
  }

  .option-chip-btn.active {
    background-color: var(--pastel-yellow);
    box-shadow: 2.5px 2.5px 0 var(--color-border);
  }

  .option-chip-btn:disabled {
    background-color: #f3f4f6;
    color: #9ca3af;
    border-color: #d1d5db;
    cursor: not-allowed;
    box-shadow: 1.5px 1.5px 0 #d1d5db;
  }

  .gps-warning-hint {
    font-size: 0.8rem;
    font-weight: 700;
    color: #dc2626;
    margin-top: 8px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .gps-status-pill {
    margin-top: 20px;
    padding: 10px 14px;
    background-color: var(--bg-secondary);
    border: var(--border-width-sm) solid var(--color-border);
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .gps-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    border: 1px solid var(--color-border);
  }

  .gps-dot.pulsing { background-color: #eab308; animation: pulse 1s infinite; }
  .gps-dot.success { background-color: #10b981; }
  .gps-dot.warning { background-color: #ef4444; }

  .gps-text {
    font-size: 0.82rem;
    font-weight: 700;
    color: var(--text-dark);
  }

  .apply-action-bar {
    margin-top: 20px;
  }

  .btn-apply-full {
    width: 100%;
    padding: 12px !important;
    font-size: 0.98rem !important;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
  }
</style>
