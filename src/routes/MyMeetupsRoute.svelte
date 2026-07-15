<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { onAuthStateChanged, type User } from "firebase/auth";
  import { auth } from "../lib/firebase";
  import MeetupList from "../lib/MeetupList.svelte";
  import Icon from "../lib/Icon.svelte";

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
    approvedUids?: string[];
  }

  interface Props {
    meetups: Meetup[];
    userLat: number | null;
    userLng: number | null;
    selectedCity: "all" | "HCM" | "HN";
    selectedDistance: "all" | "5" | "10";
    isTrackingGPS: boolean;
    gpsError: boolean;
    addToast: (msg: string, type: "success" | "error" | "info") => void;
    isLoading?: boolean;
  }

  let {
    meetups,
    userLat,
    userLng,
    selectedCity = $bindable("all"),
    selectedDistance = $bindable("all"),
    isTrackingGPS,
    gpsError,
    isLoading = false,
  }: Props = $props();

  let currentUser = $state<User | null>(auth.currentUser);
  let unsubscribeAuth: (() => void) | null = null;
  let activeTab = $state<"all" | "host" | "member" | "pending">("all");

  onMount(() => {
    unsubscribeAuth = onAuthStateChanged(auth, (user) => {
      currentUser = user;
    });
  });

  $effect(() => {
    if (auth.currentUser !== currentUser) {
      currentUser = auth.currentUser;
    }
  });

  onDestroy(() => {
    if (unsubscribeAuth) unsubscribeAuth();
  });

  // Lọc danh sách kèo liên quan tới user hiện tại
  let myMeetups = $derived.by(() => {
    if (!currentUser) return [];
    const uid = currentUser.uid;

    return meetups.filter((m) => {
      const isUserHost = m.hostUid === uid || m.host_uid === uid;
      const isUserMember = Array.isArray(m.approvedUids) && m.approvedUids.includes(uid);
      const isUserPending = (Array.isArray(m.pendingUids) && m.pendingUids.includes(uid)) ||
                            (Array.isArray(m.approvedPendingUids) && m.approvedPendingUids.includes(uid));

      if (activeTab === "all") {
        return isUserHost || isUserMember || isUserPending;
      }
      if (activeTab === "host") {
        return isUserHost;
      }
      if (activeTab === "member") {
        return isUserMember && !isUserHost;
      }
      if (activeTab === "pending") {
        return isUserPending;
      }
      return false;
    });
  });
</script>

<section id="my-meetups-route">
  <h2 class="section-title">Các Kèo Của Bạn</h2>

  {#if currentUser}
    <!-- Tab Filter Bar (Neo-brutalist Style) -->
    <div class="cartoon-card my-meetups-filter">
      <span class="filter-label">Lọc kèo:</span>
      <div class="tab-btn-group">
        <button
          type="button"
          class="tab-btn-pill {activeTab === 'all' ? 'active' : ''}"
          onclick={() => activeTab = "all"}
        >
          <Icon name="dice" size={15} />
          <span>Tất cả kèo</span>
        </button>
        <button
          type="button"
          class="tab-btn-pill {activeTab === 'host' ? 'active' : ''}"
          onclick={() => activeTab = "host"}
        >
          <Icon name="crown" size={15} />
          <span>Tôi làm Host</span>
        </button>
        <button
          type="button"
          class="tab-btn-pill {activeTab === 'member' ? 'active' : ''}"
          onclick={() => activeTab = "member"}
        >
          <Icon name="check-circle" size={15} />
          <span>Đã tham gia</span>
        </button>
        <button
          type="button"
          class="tab-btn-pill {activeTab === 'pending' ? 'active' : ''}"
          onclick={() => activeTab = "pending"}
        >
          <Icon name="clock" size={15} />
          <span>Đang chờ duyệt</span>
        </button>
      </div>
    </div>

    <!-- Render filtered meetup cards -->
    <MeetupList
      meetups={myMeetups}
      {userLat}
      {userLng}
      bind:selectedCity
      bind:selectedDistance
      {isTrackingGPS}
      {gpsError}
      showFilterBar={false}
      {isLoading}
    />
  {:else}
    <!-- Unauthorized warning -->
    <div class="cartoon-card locked-card">
      <div class="locked-icon">
        <Icon name="lock" size={40} />
      </div>
      <h4 class="locked-title">Cần Đăng Nhập Tài Khoản</h4>
      <p class="locked-description">
        Bạn cần đăng nhập để xem danh sách các kèo chơi boardgame mà bạn làm host hoặc đã đăng ký tham gia chơi!
      </p>
      <a
        href="#/profile"
        class="btn btn-primary login-link-btn"
      >
        <Icon name="key" size={16} />
        <span>Đi tới trang Đăng nhập</span>
      </a>
    </div>
  {/if}
</section>

<style>
  #my-meetups-route {
    padding-bottom: 60px;
  }

  /* Tab Filter Bar */
  .my-meetups-filter {
    margin-bottom: 24px;
    padding: 15px;
    background-color: #fffefb;
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    justify-content: center;
    align-items: center;
  }

  .filter-label {
    font-weight: 800;
    font-size: 0.95rem;
    margin-right: 8px;
    color: var(--text-dark);
  }

  .tab-btn-group {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  /* Neo-brutalist pill buttons */
  .tab-btn-pill {
    padding: 8px 16px;
    font-size: 0.85rem;
    font-weight: 700;
    font-family: var(--font-family);
    border: 3px solid #1e1e24;
    border-radius: 100px;
    cursor: pointer;
    transition: transform 0.1s ease, box-shadow 0.1s ease;
    outline: none;
    background-color: #ffffff;
    box-shadow: 3px 3px 0px #1e1e24;
  }

  .tab-btn-pill.active {
    background-color: var(--pastel-yellow, #ffe869);
    box-shadow: 2px 2px 0px #1e1e24;
    transform: translate(1px, 1px);
  }

  .tab-btn-pill:hover {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px 0px #1e1e24;
  }

  .tab-btn-pill.active:hover {
    transform: translate(1px, 1px);
    box-shadow: 2px 2px 0px #1e1e24;
  }

  .tab-btn-pill:active {
    transform: translate(3px, 3px);
    box-shadow: 0px 0px 0px #1e1e24;
  }

  /* Locked State */
  .locked-card {
    padding: 40px;
    background-color: #fffefb;
    text-align: center;
    margin-top: 20px;
  }

  .locked-icon {
    font-size: 3rem;
    display: block;
    margin-bottom: 16px;
  }

  .locked-title {
    font-size: 1.3rem;
    font-weight: 700;
    margin-bottom: 8px;
    color: var(--text-dark);
  }

  .locked-description {
    font-size: 0.95rem;
    font-weight: 500;
    color: var(--text-muted);
    max-width: 360px;
    margin: 0 auto 24px auto;
    line-height: 1.5;
  }

  .login-link-btn {
    display: inline-block;
    padding: 12px 28px;
    font-size: 0.95rem;
    font-weight: 800;
    border: 3px solid #1e1e24;
    border-radius: 8px;
    background-color: var(--pastel-yellow, #ffe869);
    color: #1e1e24;
    box-shadow: 4px 4px 0px #1e1e24;
    text-decoration: none;
  }
</style>
