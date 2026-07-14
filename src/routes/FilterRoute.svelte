<script lang="ts">
  import { goBack } from '../lib/router';

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
      <h2>🔍 Tùy Chỉnh Bộ Lọc Tìm Kiếm</h2>
      <span class="sub-title">Lọc kèo chơi theo thành phố và bán kính vị trí</span>
    </div>
  </div>

  <!-- Body Content -->
  <div class="cartoon-card filter-content-card">
    <!-- City Selection -->
    <div class="filter-section">
      <h3 class="section-title-label">🏙️ Chọn Khu Vực Thành Phố:</h3>
      <div class="options-grid">
        <button 
          type="button" 
          class="option-card-btn {tempCity === 'all' ? 'active' : ''}"
          onclick={() => tempCity = 'all'}
        >
          <span class="option-icon">🌐</span>
          <span class="option-name">Tất Cả Các Thành Phố</span>
        </button>

        <button 
          type="button" 
          class="option-card-btn {tempCity === 'HCM' ? 'active' : ''}"
          onclick={() => tempCity = 'HCM'}
        >
          <span class="option-icon">🌆</span>
          <span class="option-name">TP. Hồ Chí Minh</span>
        </button>

        <button 
          type="button" 
          class="option-card-btn {tempCity === 'HN' ? 'active' : ''}"
          onclick={() => tempCity = 'HN'}
        >
          <span class="option-icon">🏰</span>
          <span class="option-name">Hà Nội</span>
        </button>
      </div>
    </div>

    <!-- Distance Selection -->
    <div class="filter-section" style="margin-top: 32px;">
      <h3 class="section-title-label">📍 Chọn Khoảng Cách Vị Trí (GPS):</h3>
      <div class="options-grid">
        <button 
          type="button" 
          class="option-card-btn {tempDistance === 'all' ? 'active' : ''}"
          onclick={() => tempDistance = 'all'}
        >
          <span class="option-icon">🗺️</span>
          <span class="option-name">Mọi Khoảng Cách</span>
        </button>

        <button 
          type="button" 
          class="option-card-btn {tempDistance === '5' ? 'active' : ''}"
          disabled={userLat === null}
          onclick={() => tempDistance = '5'}
        >
          <span class="option-icon">⚡</span>
          <span class="option-name">Dưới 5 km</span>
        </button>

        <button 
          type="button" 
          class="option-card-btn {tempDistance === '10' ? 'active' : ''}"
          disabled={userLat === null}
          onclick={() => tempDistance = '10'}
        >
          <span class="option-icon">🚀</span>
          <span class="option-name">Dưới 10 km</span>
        </button>
      </div>

      {#if userLat === null}
        <p class="gps-warning-hint">
          ⚠️ Vui lòng cho phép quyền vị trí (GPS) để bật lọc khoảng cách theo km.
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
        <span class="gps-text">Đã bật định vị GPS thành công</span>
      {:else if gpsError}
        <span class="gps-dot warning"></span>
        <span class="gps-text">GPS bị từ chối hoặc thiết bị không hỗ trợ.</span>
      {/if}
    </div>

    <div class="apply-action-bar">
      <button type="button" class="btn btn-primary btn-apply-full" onclick={handleApply}>
        Áp dụng bộ lọc & Xem kết quả 🎯
      </button>
    </div>
  </div>
</div>

<style>
  .fullscreen-route-view {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding-bottom: 40px;
  }

  .route-top-nav {
    background-color: var(--pastel-yellow);
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px 24px;
    border-radius: var(--radius-lg);
    box-shadow: 4px 4px 0 var(--color-border);
    text-align: left;
  }

  .back-btn {
    padding: 8px 16px !important;
    font-size: 0.95rem !important;
    white-space: nowrap;
  }

  .nav-title-group h2 {
    font-size: 1.3rem;
    font-weight: 800;
    margin: 0;
  }

  .sub-title {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text-dark);
  }

  .filter-content-card {
    background-color: #fffefb;
    padding: 32px 24px;
    text-align: left;
  }

  .section-title-label {
    font-size: 1.1rem;
    font-weight: 800;
    margin-bottom: 16px;
    color: var(--text-dark);
  }

  .options-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
  }

  .option-card-btn {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 10px;
    padding: 20px 16px;
    border-radius: var(--radius-md);
    border: var(--border-width) solid var(--color-border);
    background-color: #fff;
    cursor: pointer;
    box-shadow: 4px 4px 0 var(--color-border);
    font-family: var(--font-family);
    transition: all 0.15s ease;
  }

  .option-card-btn:hover:not(:disabled) {
    transform: translate(-2px, -2px);
    box-shadow: 6px 6px 0 var(--color-border);
  }

  .option-card-btn:active:not(:disabled) {
    transform: translate(2px, 2px);
    box-shadow: 0 0 0 var(--color-border);
  }

  .option-card-btn.active {
    background-color: var(--pastel-yellow);
    box-shadow: 5px 5px 0 var(--color-border);
    transform: translate(-1px, -1px);
  }

  .option-card-btn:disabled {
    background-color: #f3f4f6;
    color: #9ca3af;
    border-color: #d1d5db;
    cursor: not-allowed;
    box-shadow: 2px 2px 0 #d1d5db;
  }

  .option-icon {
    font-size: 2.2rem;
  }

  .option-name {
    font-size: 1rem;
    font-weight: 800;
  }

  .gps-warning-hint {
    font-size: 0.85rem;
    font-weight: 700;
    color: #dc2626;
    margin-top: 10px;
  }

  .gps-status-pill {
    margin-top: 24px;
    padding: 12px 16px;
    background-color: var(--bg-secondary);
    border: var(--border-width-sm) solid var(--color-border);
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .gps-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    border: 1px solid var(--color-border);
  }

  .gps-dot.pulsing { background-color: #eab308; animation: pulse 1s infinite; }
  .gps-dot.success { background-color: #10b981; }
  .gps-dot.warning { background-color: #ef4444; }

  .gps-text {
    font-size: 0.88rem;
    font-weight: 700;
    color: var(--text-dark);
  }

  .apply-action-bar {
    margin-top: 32px;
  }

  .btn-apply-full {
    width: 100%;
    padding: 16px !important;
    font-size: 1.1rem !important;
  }
</style>
