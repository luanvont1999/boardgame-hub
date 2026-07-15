<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { onAuthStateChanged, type User } from "firebase/auth";
  import { auth } from "./firebase";
  import {
    isApprovedMember,
    isHost,
    requestToJoin,
    cancelJoinRequest,
    kickOrLeaveMember,
    confirmParticipation,
  } from "./meetupService";
  import { navigate } from "./router.svelte";

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
    hostFcmToken?: string;
    pendingUids?: string[];
    approvedPendingUids?: string[];
  }

  // Props — no more callback props for chat/manage/map
  interface Props {
    meetups: Meetup[];
    userLat: number | null;
    userLng: number | null;
    selectedCity: "all" | "HCM" | "HN";
    selectedDistance: "all" | "5" | "10";
    isTrackingGPS: boolean;
    gpsError: boolean;
  }

  let {
    meetups,
    userLat,
    userLng,
    selectedCity = $bindable("all"),
    selectedDistance = $bindable("all"),
    isTrackingGPS,
    gpsError,
  }: Props = $props();

  let currentUser = $state<User | null>(auth.currentUser);
  let unsubscribeAuth: (() => void) | null = null;
  let isActionProcessing = $state<boolean>(false);

  onMount(() => {
    unsubscribeAuth = onAuthStateChanged(auth, (user) => {
      currentUser = user;
    });
  });

  onDestroy(() => {
    if (unsubscribeAuth) unsubscribeAuth();
  });  async function handleJoinRequest(meetup: Meetup) {
    if (!currentUser || isActionProcessing) return;
    isActionProcessing = true;
    try {
      await requestToJoin(meetup.id, currentUser);
    } catch (err) {
      console.error("Request join error:", err);
    } finally {
      isActionProcessing = false;
    }
  }

  async function handleCancel(meetup: Meetup) {
    if (!currentUser || isActionProcessing) return;
    isActionProcessing = true;
    try {
      await cancelJoinRequest(meetup.id, currentUser.uid);
    } catch (err) {
      console.error("Cancel request error:", err);
    } finally {
      isActionProcessing = false;
    }
  }

  async function handleLeave(meetup: Meetup) {
    if (!currentUser || isActionProcessing) return;
    isActionProcessing = true;
    try {
      const name = currentUser.displayName || currentUser.email || 'Thành viên';
      await kickOrLeaveMember(meetup.id, currentUser.uid, name, false);
    } catch (err) {
      console.error("Leave meetup error:", err);
    } finally {
      isActionProcessing = false;
    }
  }

  async function handleConfirm(meetup: Meetup) {
    if (!currentUser || isActionProcessing) return;
    isActionProcessing = true;
    try {
      const name = currentUser.displayName || currentUser.email || 'Thành viên';
      await confirmParticipation(meetup.id, currentUser.uid, name);
    } catch (err) {
      console.error("Confirm participation error:", err);
    } finally {
      isActionProcessing = false;
    }
  }

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

  function getMeetupCity(meetup: Meetup): "HCM" | "HN" {
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
      return `${hours}:${minutes}, ${day}/${month}/${year}`;
    } catch {
      return timeStr;
    }
  }

  // Navigate instead of opening local modal for filter and map
  function openFilter() {
    navigate({ name: "filter" });
  }
  function openMap() {
    navigate({ name: "map", mode: "discover" });
  }
</script>

<div class="meetup-list-container">
  <!-- Cartoon Trigger Bar for Search Filter Modal -->
  <div class="cartoon-card filter-trigger-card">
    <div class="filter-trigger-info">
      <div class="filter-trigger-header">
        <span class="filter-title-icon">🎲</span>
        <h3>Danh Sách Kèo Chơi Boardgame</h3>
      </div>

      <!-- Active Filter Badges -->
      <div class="active-filter-badges">
        <span class="badge-chip">
          📍 Khu vực: <strong
            >{selectedCity === "all"
              ? "Tất cả"
              : selectedCity === "HCM"
                ? "TP. HCM"
                : "Hà Nội"}</strong
          >
        </span>
        <span class="badge-chip">
          📏 Khoảng cách: <strong
            >{selectedDistance === "all"
              ? "Mọi khoảng cách"
              : `< ${selectedDistance} km`}</strong
          >
        </span>
        {#if userLat !== null}
          <span class="badge-chip gps-on-chip"> 🎯 GPS Bật </span>
        {/if}
      </div>
    </div>

    <button
      type="button"
      class="btn btn-primary filter-trigger-btn"
      onclick={openFilter}
    >
      🔍 Bộ Lọc Tìm Kiếm ⚙️
    </button>
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
        {@const distance =
          userLat !== null && userLng !== null
            ? calculateDistance(userLat, userLng, meetup.lat, meetup.lng)
            : null}
        {@const userIsHost = isHost(meetup, currentUser?.uid)}
        {@const userIsMember = isApprovedMember(meetup, currentUser?.uid)}
        {@const userIsPending =
          Array.isArray(meetup.pendingUids) &&
          meetup.pendingUids.includes(currentUser?.uid || "")}
        {@const userIsApprovedPending =
          Array.isArray(meetup.approvedPendingUids) &&
          meetup.approvedPendingUids.includes(currentUser?.uid || "")}

        <div
          class="cartoon-card meetup-item-card"
          style="border-top: 10px solid {meetup.color};"
        >
          <!-- Card Badge info -->
          <div class="meetup-card-badges">
            <span
              class="city-badge {city === 'HCM' ? 'hcm-color' : 'hn-color'}"
            >
              {city === 'HCM' ? '🌆 HCM' : '🏰 HN'}
            </span>
            {#if distance !== null}
              <span class="distance-badge">
                📍 Cách {distance.toFixed(1)} km
              </span>
            {/if}
            {#if userIsHost}
              <span class="membership-badge host-tag">👑 Host</span>
              {#if Array.isArray(meetup.pendingUids) && meetup.pendingUids.length > 0}
                <span class="membership-badge pending-alert-tag">🔔 {meetup.pendingUids.length} mới</span>
              {/if}
            {:else if userIsMember}
              <span class="membership-badge member-tag">✅ Đã tham gia</span>
            {:else if userIsApprovedPending}
              <span class="membership-badge approved-pending-tag pulsing-badge">🎉 Được duyệt</span>
            {:else if userIsPending}
              <span class="membership-badge pending-tag">⏳ Chờ duyệt</span>
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
              <span
                >Sĩ số: <strong
                  >{meetup.players_count || meetup.playersCount || 1} / {meetup.players_needed ||
                    meetup.playersNeeded ||
                    4} người</strong
                ></span
              >
            </div>
            <div class="detail-row">
              <span class="detail-icon">👤</span>
              <span
                >Host: <strong
                  >{meetup.host_name || meetup.hostName || "Ẩn danh"}</strong
                ></span
              >
            </div>
          </div>

          <div class="card-action-row">
            <button
              class="btn btn-secondary action-btn"
              onclick={() => navigate({ name: "map", mode: "discover" })}
            >
              Vị trí 🗺️
            </button>

            {#if userIsMember}
              <button
                class="btn btn-primary action-btn"
                onclick={() => navigate({ name: "chat", meetup })}
              >
                Chat 💬
              </button>
            {/if}

            {#if userIsHost}
              <button
                class="btn btn-success action-btn"
                onclick={() => navigate({ name: "manage", meetup })}
              >
                Duyệt 👥
                {#if Array.isArray(meetup.pendingUids) && meetup.pendingUids.length > 0}
                  <span class="btn-alert-dot"></span>
                {/if}
              </button>
            {:else if userIsMember}
              <button
                class="btn btn-secondary action-btn leave-btn"
                onclick={() => handleLeave(meetup)}
                disabled={isActionProcessing}
              >
                Rời 🚪
              </button>
            {:else if userIsApprovedPending}
              <button
                class="btn btn-warning action-btn pulse-button"
                onclick={() => handleConfirm(meetup)}
                disabled={isActionProcessing}
              >
                {isActionProcessing ? "..." : "Xác nhận 🤝"}
              </button>
              <button
                class="btn btn-secondary action-btn cancel-btn"
                onclick={() => handleCancel(meetup)}
                disabled={isActionProcessing}
              >
                Từ chối ✕
              </button>
            {:else if userIsPending}
              <button
                class="btn btn-secondary action-btn cancel-btn"
                onclick={() => handleCancel(meetup)}
                disabled={isActionProcessing}
              >
                {isActionProcessing ? "..." : "Hủy ✕"}
              </button>
            {:else if currentUser}
              <button
                class="btn btn-success action-btn"
                onclick={() => handleJoinRequest(meetup)}
                disabled={isActionProcessing}
              >
                {isActionProcessing ? "..." : "Xin vào 🤝"}
              </button>
            {/if}
          </div>
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

  .filter-trigger-card {
    background-color: #fffefb;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 20px 24px;
    flex-wrap: wrap;
  }

  @media (max-width: 600px) {
    .filter-trigger-card {
      flex-direction: column;
      align-items: stretch;
    }
    .filter-trigger-btn {
      width: 100%;
    }
  }

  .filter-trigger-info {
    display: flex;
    flex-direction: column;
    gap: 8px;
    text-align: left;
  }

  .filter-trigger-header {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .filter-trigger-header h3 {
    font-size: 1.3rem;
    font-weight: 800;
  }

  .active-filter-badges {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .badge-chip {
    font-size: 0.85rem;
    font-weight: 600;
    padding: 4px 12px;
    background-color: var(--bg-secondary);
    border: var(--border-width-sm) solid var(--color-border);
    border-radius: 100px;
    color: var(--text-dark);
    box-shadow: 1.5px 1.5px 0 var(--color-border);
  }

  .gps-on-chip {
    background-color: var(--pastel-green);
  }

  .filter-trigger-btn {
    font-size: 1rem;
    padding: 10px 20px;
    white-space: nowrap;
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

  .city-badge,
  .distance-badge {
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

  .card-action-row {
    width: 100%;
    display: flex;
    gap: 10px;
    margin-top: auto;
  }

  .action-btn {
    flex: 1;
    padding: 10px 12px !important;
    font-size: 0.9rem !important;
    white-space: nowrap;
  }

  .leave-btn {
    background-color: var(--pastel-pink) !important;
  }

  .membership-badge {
    font-size: 0.8rem;
    font-weight: 800;
    padding: 4px 10px;
    border-radius: 100px;
    border: var(--border-width-sm) solid var(--color-border);
    box-shadow: 1.5px 1.5px 0 var(--color-border);
  }

  .host-tag {
    background-color: var(--pastel-yellow);
  }

  .member-tag {
    background-color: var(--pastel-green);
  }

  .pending-tag {
    background-color: var(--pastel-orange, #ffc078);
  }

  .cancel-btn {
    background-color: #fca5a5 !important; /* màu đỏ pastel nhẹ nhàng */
  }

  .pending-alert-tag {
    background-color: #ef4444 !important;
    color: white !important;
    animation: wiggle 0.5s ease-in-out infinite alternate;
  }

  .btn-alert-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    background-color: #ef4444;
    border-radius: 50%;
    margin-left: 4px;
    border: 1px solid white;
    vertical-align: middle;
    animation: pulse-red 1s infinite;
  }

  @keyframes pulse-red {
    0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7); }
    70% { transform: scale(1); box-shadow: 0 0 0 4px rgba(239, 68, 68, 0); }
    100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(239, 68, 68, 0); }
  }

  @keyframes wiggle {
    0% { transform: rotate(-2deg); }
    100% { transform: rotate(2deg); }
  }

  .approved-pending-tag {
    background-color: #fef08a !important;
    color: var(--text-dark) !important;
  }

  .pulsing-badge {
    animation: wiggle 0.5s ease-in-out infinite alternate;
  }

  .pulse-button {
    background-color: var(--pastel-yellow, #fef08a) !important;
    animation: pulse-glow 1.5s infinite;
  }

  @keyframes pulse-glow {
    0% {
      transform: scale(1);
    }
    50% {
      transform: scale(1.05);
    }
    100% {
      transform: scale(1);
    }
  }
</style>
