<script lang="ts">
  import { onMount } from 'svelte';
  import { onAuthStateChanged, type User } from 'firebase/auth';
  import { collection, addDoc, serverTimestamp } from 'firebase/firestore';
  import { auth, db } from './firebase';


  // Svelte 5 props with bindable selectedLat, selectedLng, addressText and proximity current GPS
  interface Props {
    selectedLat: number | null;
    selectedLng: number | null;
    userLat?: number | null;
    userLng?: number | null;
    addressText?: string;
    onCreateSuccess: () => void;
    onLocationInputClick?: () => void;
  }
  let { 
    selectedLat = $bindable(null), 
    selectedLng = $bindable(null), 
    userLat = null,
    userLng = null,
    addressText = $bindable(''),
    onCreateSuccess, 
    onLocationInputClick 
  }: Props = $props();

  // State variables (Svelte 5 runes)
  let user = $state<User | null>(null);
  let title = $state<string>('');
  let game = $state<string>('');
  let time = $state<string>(''); // datetime-local format
  let lat = $state<number | null>(null);
  let lng = $state<number | null>(null);

  // Address Geocoding states
  let suggestions = $state<any[]>([]);
  let showDropdown = $state<boolean>(false);
  let isSearchingAddress = $state<boolean>(false);
  let debounceTimer: any = null;

  // Preset popular board game cafes
  const popularVenues = [
    { name: "🎲 Boardgame Station", address: "21 Cô Bắc, Quận 1, TP. HCM", lat: 10.7656, lng: 106.6961 },
    { name: "☕ Cashflow Cafe", address: "7A/19 Thành Thái, Quận 10, TP. HCM", lat: 10.7712, lng: 106.6644 },
    { name: "🐺 Ma Sói Guild", address: "12 Vệ Hồ, Tây Hồ, Hà Nội", lat: 21.0620, lng: 105.8155 },
    { name: "🏰 The Boardgame Hub", address: "45 Lương Ngọc Quyến, Hoàn Kiếm, Hà Nội", lat: 21.0345, lng: 105.8524 }
  ];

  let isSubmitting = $state<boolean>(false);
  let errorMessage = $state<string>('');
  let successMessage = $state<string>('');

  async function reverseGeocode(targetLat: number, targetLng: number) {
    const token = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;
    if (!token) return;

    isSearchingAddress = true;
    const url = `https://api.mapbox.com/geocoding/v5/mapbox.places/${targetLng},${targetLat}.json?access_token=${token}&country=vn&limit=1&language=vi`;

    try {
      const res = await fetch(url);
      if (res.ok) {
        const data = await res.json();
        if (data.features && data.features.length > 0) {
          addressText = data.features[0].place_name;
        } else {
          addressText = `Tọa độ: ${targetLat.toFixed(5)}, ${targetLng.toFixed(5)}`;
        }
      } else {
        addressText = `Tọa độ: ${targetLat.toFixed(5)}, ${targetLng.toFixed(5)}`;
      }
    } catch (err) {
      console.error("Reverse geocoding failed:", err);
      addressText = `Tọa độ: ${targetLat.toFixed(5)}, ${targetLng.toFixed(5)}`;
    } finally {
      isSearchingAddress = false;
    }
  }

  // Sync selected location from Map click prop
  $effect(() => {
    if (selectedLat !== null && selectedLng !== null) {
      if (lat !== selectedLat || lng !== selectedLng) {
        lat = selectedLat;
        lng = selectedLng;
        // Resolve coordinates to a readable address name
        reverseGeocode(selectedLat, selectedLng);
      }
    } else {
      // Only clear if lat/lng are active
      if (lat !== null || lng !== null) {
        lat = null;
        lng = null;
        addressText = '';
      }
    }
  });

  onMount(() => {
    const unsubscribe = onAuthStateChanged(auth, (currentUser) => {
      user = currentUser;
    });
    return () => {
      unsubscribe();
      clearTimeout(debounceTimer);
    };
  });

  async function searchAddress() {
    if (addressText.trim().length < 3) {
      suggestions = [];
      return;
    }

    const token = import.meta.env.VITE_MAPBOX_ACCESS_TOKEN;
    if (!token) return;

    isSearchingAddress = true;
    
    // Add proximity bias using user coordinates if GPS is available
    let proximityQuery = '';
    if (userLat !== null && userLng !== null) {
      proximityQuery = `&proximity=${userLng},${userLat}`;
    }

    // Set types=poi,address to search specifically for cafes, businesses, or streets
    const url = `https://api.mapbox.com/geocoding/v5/mapbox.places/${encodeURIComponent(addressText)}.json?access_token=${token}&country=vn&limit=5&language=vi&types=poi,address${proximityQuery}`;

    try {
      const res = await fetch(url);
      if (res.ok) {
        const data = await res.json();
        if (addressText.trim().length >= 3) {
          suggestions = data.features || [];
        }
      }
    } catch (err) {
      console.error("Geocoding lookup failed:", err);
    } finally {
      isSearchingAddress = false;
    }
  }

  function handleAddressInput() {
    showDropdown = true;
    
    // Clear previous coordinate selection so they must pick one
    lat = null;
    lng = null;
    selectedLat = null;
    selectedLng = null;

    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      searchAddress();
    }, 400);
  }

  function selectSuggestion(item: any) {
    const [itemLng, itemLat] = item.center;
    
    // Update internal coordinates
    lat = itemLat;
    lng = itemLng;
    
    // Bind coordinates to parent props (will update Map and App parent)
    selectedLat = itemLat;
    selectedLng = itemLng;

    addressText = item.place_name;
    suggestions = [];
    showDropdown = false;
  }

  function selectPopularVenue(venue: typeof popularVenues[0]) {
    lat = venue.lat;
    lng = venue.lng;
    selectedLat = venue.lat;
    selectedLng = venue.lng;
    addressText = `${venue.name} (${venue.address})`;
    suggestions = [];
    showDropdown = false;
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!user) {
      errorMessage = 'Bạn phải đăng nhập để lên kèo chơi!';
      return;
    }

    if (!title || !game || !time || lat === null || lng === null) {
      errorMessage = 'Vui lòng điền đầy đủ thông tin và nhập/chọn địa chỉ!';
      return;
    }

    errorMessage = '';
    successMessage = '';
    isSubmitting = true;

    try {
      const colors = ["#bca0f5", "#ffa4b2", "#ffe869", "#ffb875", "#9ee3b2", "#a4f0fd"];
      const color = colors[Math.floor(Math.random() * colors.length)];
      const hostName = user.displayName || user.email || 'Ẩn danh';

      // Save meetup directly to Firebase Cloud Firestore
      await addDoc(collection(db, 'meetups'), {
        title,
        game,
        hostName,
        hostUid: user.uid,
        lat,
        lng,
        playersCount: 1,
        playersNeeded: 4,
        time,
        color,
        createdAt: serverTimestamp()
      });

      successMessage = 'Tạo kèo mới thành công! Quân meeple của bạn đã được ghim lên bản đồ.';
      
      // Reset form fields
      title = '';
      game = '';
      time = '';
      addressText = '';
      lat = null;
      lng = null;
      selectedLat = null;
      selectedLng = null;

      onCreateSuccess();
    } catch (err: any) {
      console.error('[Firestore] Create meetup failed:', err);
      errorMessage = 'Lỗi tạo kèo trên Firestore: ' + (err.message || 'Lỗi kết nối');
    } finally {
      isSubmitting = false;
    }
  }

  // Close dropdown if clicking outside
  function handleGlobalClick(e: MouseEvent) {
    const target = e.target as HTMLElement;
    const isInsideSearch = target.closest('.location-input-container') || target.closest('.autocomplete-dropdown-list');
    if (!isInsideSearch) {
      showDropdown = false;
    }
  }
</script>

<svelte:window onclick={handleGlobalClick} />

<div class="create-meetup-wrapper">
  {#if !user}
    <!-- Locked State Warning -->
    <div class="cartoon-card locked-card">
      <div class="lock-header">
        <span class="lock-icon">🔒</span>
        <h3>Lên Kèo Đang Bị Khóa</h3>
      </div>
      <p style="margin-bottom: 15px; font-weight: 500;">
        Bạn cần đăng nhập tài khoản Firebase để có thể tự lên kèo chơi riêng và mời mọi người tham gia.
      </p>
      <a href="#auth" class="btn btn-primary">Đi đến Đăng nhập ngay</a>
    </div>
  {:else}
    <!-- Playful Cartoon Form -->
    <div class="cartoon-card form-card">
      <h3 class="form-title">🎲 Lên Kèo Chơi Mới</h3>
      
      {#if errorMessage}
        <div class="alert error-alert">{errorMessage}</div>
      {/if}
      {#if successMessage}
        <div class="alert success-alert">{successMessage}</div>
      {/if}

      <form onsubmit={handleSubmit} class="meetup-form">
        <div class="form-group">
          <label for="meetup-title">Tên Kèo Chơi:</label>
          <input 
            type="text" 
            id="meetup-title" 
            placeholder="Ví dụ: Đại chiến Werewolf, Tìm chân Catan..." 
            bind:value={title}
            disabled={isSubmitting}
            required 
          />
        </div>

        <div class="form-group">
          <label for="meetup-game">Các Boardgame Sẽ Chơi:</label>
          <input 
            type="text" 
            id="meetup-game" 
            placeholder="Ví dụ: Catan, Ma sói, Avalon, Mèo nổ..." 
            bind:value={game}
            disabled={isSubmitting}
            required 
          />
        </div>

        <div class="form-group-row">
          <div class="form-group">
            <label for="meetup-time">Thời Gian:</label>
            <input 
              type="datetime-local" 
              id="meetup-time" 
              bind:value={time}
              disabled={isSubmitting}
              required 
            />
          </div>
        </div>

        <div class="form-group" style="position: relative;">
          <label for="meetup-location">Vị Trí Kèo Chơi:</label>
          
          <div class="location-input-container">
            <input 
              type="text" 
              id="meetup-location" 
              placeholder="Nhập địa chỉ hoặc bấm nút Bản đồ..." 
              bind:value={addressText}
              oninput={handleAddressInput}
              onfocus={() => showDropdown = true}
              disabled={isSubmitting}
              autocomplete="off"
              required 
            />
            <button 
              type="button" 
              class="btn btn-secondary btn-map-select" 
              onclick={onLocationInputClick}
              disabled={isSubmitting}
            >
              🗺️ Bản đồ
            </button>
          </div>

          <!-- Indication badge showing if coordinates are captured -->
          {#if lat !== null && lng !== null}
            <span class="location-status-badge">✓ Đã nhận tọa độ</span>
          {:else if isSearchingAddress}
            <span class="location-status-badge searching">Đang quét địa điểm...</span>
          {/if}

          <!-- Popular preset locations -->
          <div class="popular-venues-container">
            <span class="popular-title">💡 Gợi ý nhanh điểm chơi boardgame:</span>
            <div class="venue-tags">
              {#each popularVenues as venue}
                <button 
                  type="button" 
                  class="venue-tag-btn" 
                  onclick={() => selectPopularVenue(venue)}
                  disabled={isSubmitting}
                >
                  {venue.name.split(" - ")[0]}
                </button>
              {/each}
            </div>
          </div>

          <!-- Autocomplete Dropdown List -->
          {#if showDropdown && suggestions.length > 0}
            <ul class="autocomplete-dropdown-list cartoon-card">
              {#each suggestions as item}
                <li class="dropdown-item">
                  <button 
                    type="button" 
                    onclick={() => selectSuggestion(item)}
                  >
                    📍 {item.place_name}
                  </button>
                </li>
              {/each}
            </ul>
          {/if}
        </div>

        <button type="submit" class="btn btn-primary w-full" style="margin-top: 10px;" disabled={isSubmitting}>
          {isSubmitting ? 'Đang tạo kèo...' : 'Lên kèo chơi ngay! 🚀'}
        </button>
      </form>
    </div>
  {/if}
</div>

<style>
  .create-meetup-wrapper {
    width: 100%;
    max-width: 480px;
    margin: 0 auto;
  }

  .form-card {
    background-color: var(--pastel-yellow);
    padding: 24px;
    text-align: left;
  }

  .form-title {
    font-size: 1.5rem;
    margin-bottom: 20px;
    color: var(--text-dark);
    text-align: center;
  }

  .meetup-form {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .form-group label {
    font-weight: 800;
    font-size: 0.95rem;
    color: var(--text-dark);
  }

  .form-group input {
    padding: 12px 14px;
    border-radius: var(--radius-md);
    border: var(--border-width) solid var(--color-border);
    font-family: var(--font-family);
    font-size: 0.95rem;
    font-weight: 600;
    box-shadow: 2px 2px 0 var(--color-border);
    outline: none;
    background-color: #fff;
    transition: all 0.1s ease;
  }

  .form-group input:focus {
    transform: translate(-1px, -1px);
    box-shadow: 3px 3px 0 var(--color-border);
  }

  /* Location input & map button side by side */
  .location-input-container {
    display: flex;
    gap: 8px;
    width: 100%;
  }

  .location-input-container input {
    flex: 1;
    min-width: 0;
  }

  .btn-map-select {
    padding: 10px 14px !important;
    font-size: 0.9rem !important;
    white-space: nowrap;
    border-radius: var(--radius-md) !important;
    box-shadow: 2px 2px 0 var(--color-border) !important;
  }

  /* Location status indicator badge */
  .location-status-badge {
    align-self: flex-start;
    font-size: 0.75rem;
    font-weight: 800;
    color: #10b981; /* green */
    margin-top: 2px;
  }

  .location-status-badge.searching {
    color: #eab308; /* yellow */
  }

  /* Popular preset locations tags styling */
  .popular-venues-container {
    margin-top: 4px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .popular-title {
    font-size: 0.75rem;
    font-weight: 800;
    color: var(--text-muted);
  }

  .venue-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .venue-tag-btn {
    padding: 4px 10px;
    font-size: 0.7rem;
    font-weight: 800;
    border-radius: 8px;
    border: var(--border-width-sm) solid var(--color-border);
    background-color: #fff;
    box-shadow: 1.5px 1.5px 0 var(--color-border);
    cursor: pointer;
    font-family: var(--font-family);
    transition: all 0.1s ease;
  }

  .venue-tag-btn:hover:not(:disabled) {
    transform: translate(-1px, -1px);
    box-shadow: 2px 2px 0 var(--color-border);
    background-color: var(--pastel-pink);
  }

  .venue-tag-btn:active:not(:disabled) {
    transform: translate(1px, 1px);
    box-shadow: 0 0 0 var(--color-border);
  }

  .venue-tag-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  /* Autocomplete Dropdown List */
  .autocomplete-dropdown-list {
    position: absolute;
    top: calc(100% - 4px);
    left: 0;
    right: 0;
    z-index: 1100;
    background-color: #fff;
    border: var(--border-width) solid var(--color-border) !important;
    border-radius: var(--radius-md) !important;
    box-shadow: 4px 4px 0 var(--color-border) !important;
    list-style: none;
    margin: 4px 0 0 0;
    padding: 6px 0;
    max-height: 220px;
    overflow-y: auto;
  }

  .dropdown-item {
    width: 100%;
  }

  .dropdown-item button {
    background: none;
    border: none;
    width: 100%;
    text-align: left;
    padding: 10px 14px;
    font-weight: 700;
    font-family: var(--font-family);
    font-size: 0.85rem;
    cursor: pointer;
    display: block;
    color: var(--text-dark);
    transition: background-color 0.1s;
    line-height: 1.4;
  }

  .dropdown-item button:hover {
    background-color: var(--pastel-pink);
  }

  /* Alerts */
  .alert {
    padding: 10px 14px;
    border-radius: var(--radius-md);
    border: var(--border-width-sm) solid var(--color-border);
    font-weight: 700;
    font-size: 0.9rem;
    margin-bottom: 16px;
  }

  .error-alert {
    background-color: #ffccd3;
    color: #b91c1c;
  }

  .success-alert {
    background-color: #d1fae5;
    color: #065f46;
  }

  /* Locked card styling */
  .locked-card {
    background-color: #ffebd1;
    border-color: var(--color-border);
    padding: 30px 24px;
    text-align: center;
  }

  .lock-header {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    margin-bottom: 15px;
    color: #92400e;
  }

  .lock-icon {
    font-size: 1.8rem;
  }

  .w-full {
    width: 100%;
  }
</style>
