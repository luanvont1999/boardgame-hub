<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { onAuthStateChanged, type User } from "firebase/auth";
  import { auth } from "../lib/firebase";
  import MeetupList from "../lib/MeetupList.svelte";

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
  let activeTab = $state<"all" | "host" | "member" | "pending">("all");

  onMount(() => {
    unsubscribeAuth = onAuthStateChanged(auth, (user) => {
      currentUser = user;
    });
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
        return isUserMember && !isUserHost; // Thành viên nhưng không phải host
      }
      if (activeTab === "pending") {
        return isUserPending;
      }
      return false;
    });
  });
</script>

<section id="my-meetups-route" style="padding-bottom: 60px;">
  <h2 class="section-title">Các Kèo Của Bạn</h2>

  {#if currentUser}
    <!-- Tab Filter Bar (Neo-brutalist Style) -->
    <div class="cartoon-card my-meetups-filter" style="margin-bottom: 24px; padding: 15px; background-color: #fffefb; display: flex; flex-wrap: wrap; gap: 10px; justify-content: center; align-items: center;">
      <span style="font-weight: 800; font-size: 0.95rem; margin-right: 8px; color: var(--text-dark);">Lọc kèo:</span>
      <div style="display: flex; flex-wrap: wrap; gap: 8px;">
        <button
          type="button"
          class="tab-btn-pill {activeTab === 'all' ? 'active' : ''}"
          onclick={() => activeTab = "all"}
          style="padding: 8px 16px; font-size: 0.85rem; font-weight: 700; border: 3px solid #1e1e24; border-radius: 100px; cursor: pointer; transition: transform 0.1s ease, box-shadow 0.1s ease; outline: none;
                 background-color: {activeTab === 'all' ? 'var(--pastel-yellow, #ffe869)' : '#ffffff'};
                 box-shadow: {activeTab === 'all' ? '2px 2px 0px #1e1e24' : '3px 3px 0px #1e1e24'};
                 transform: {activeTab === 'all' ? 'translate(1px, 1px)' : 'none'};"
        >
          Tất cả kèo 🎲
        </button>
        <button
          type="button"
          class="tab-btn-pill {activeTab === 'host' ? 'active' : ''}"
          onclick={() => activeTab = "host"}
          style="padding: 8px 16px; font-size: 0.85rem; font-weight: 700; border: 3px solid #1e1e24; border-radius: 100px; cursor: pointer; transition: transform 0.1s ease, box-shadow 0.1s ease; outline: none;
                 background-color: {activeTab === 'host' ? 'var(--pastel-yellow, #ffe869)' : '#ffffff'};
                 box-shadow: {activeTab === 'host' ? '2px 2px 0px #1e1e24' : '3px 3px 0px #1e1e24'};
                 transform: {activeTab === 'host' ? 'translate(1px, 1px)' : 'none'};"
        >
          Tôi làm Host 👑
        </button>
        <button
          type="button"
          class="tab-btn-pill {activeTab === 'member' ? 'active' : ''}"
          onclick={() => activeTab = "member"}
          style="padding: 8px 16px; font-size: 0.85rem; font-weight: 700; border: 3px solid #1e1e24; border-radius: 100px; cursor: pointer; transition: transform 0.1s ease, box-shadow 0.1s ease; outline: none;
                 background-color: {activeTab === 'member' ? 'var(--pastel-yellow, #ffe869)' : '#ffffff'};
                 box-shadow: {activeTab === 'member' ? '2px 2px 0px #1e1e24' : '3px 3px 0px #1e1e24'};
                 transform: {activeTab === 'member' ? 'translate(1px, 1px)' : 'none'};"
        >
          Đã tham gia ✅
        </button>
        <button
          type="button"
          class="tab-btn-pill {activeTab === 'pending' ? 'active' : ''}"
          onclick={() => activeTab = "pending"}
          style="padding: 8px 16px; font-size: 0.85rem; font-weight: 700; border: 3px solid #1e1e24; border-radius: 100px; cursor: pointer; transition: transform 0.1s ease, box-shadow 0.1s ease; outline: none;
                 background-color: {activeTab === 'pending' ? 'var(--pastel-yellow, #ffe869)' : '#ffffff'};
                 box-shadow: {activeTab === 'pending' ? '2px 2px 0px #1e1e24' : '3px 3px 0px #1e1e24'};
                 transform: {activeTab === 'pending' ? 'translate(1px, 1px)' : 'none'};"
        >
          Đang chờ duyệt ⏳
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
    />
  {:else}
    <!-- Unauthorized warning -->
    <div class="cartoon-card" style="padding: 40px; background-color: #fffefb; text-align: center; margin-top: 20px;">
      <span style="font-size: 3rem; display: block; margin-bottom: 16px;">🔒</span>
      <h4 style="font-size: 1.3rem; font-weight: 700; margin-bottom: 8px; color: var(--text-dark);">Cần Đăng Nhập Tài Khoản</h4>
      <p style="font-size: 0.95rem; font-weight: 500; color: var(--text-muted); max-width: 360px; margin: 0 auto 24px auto; line-height: 1.5;">
        Bạn cần đăng nhập để xem danh sách các kèo chơi boardgame mà bạn làm host hoặc đã đăng ký tham gia chơi!
      </p>
      <a 
        href="#/profile" 
        class="btn btn-primary"
        style="display: inline-block; padding: 12px 28px; font-size: 0.95rem; font-weight: 800; border: 3px solid #1e1e24; border-radius: 8px; background-color: var(--pastel-yellow, #ffe869) !important; color: #1e1e24 !important; box-shadow: 4px 4px 0px #1e1e24; text-decoration: none;"
      >
        Đi tới trang Đăng nhập 🔑
      </a>
    </div>
  {/if}
</section>

<style>
  /* Hover effects for custom Brutalism pills */
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
</style>
