<script lang="ts">
  import { onMount } from 'svelte';
  import { auth } from '../lib/firebase';
  import { onAuthStateChanged, type User } from 'firebase/auth';
  import CreateMeetupForm from '../lib/CreateMeetupForm.svelte';
  import MeetupList from '../lib/MeetupList.svelte';

  interface Props {
    meetups?: any[];
    selectedLat: number | null;
    selectedLng: number | null;
    addressText: string;
    userLat: number | null;
    userLng: number | null;
  }

  let {
    meetups = [],
    selectedLat = $bindable(null),
    selectedLng = $bindable(null),
    addressText = $bindable(''),
    userLat,
    userLng,
  }: Props = $props();

  let currentUser = $state<User | null>(auth.currentUser);
  let activeTab = $state<'new' | 'active'>('new');

  onMount(() => {
    const unsub = onAuthStateChanged(auth, (user) => {
      currentUser = user;
    });
    return unsub;
  });

  // Kèo của tôi: do mình làm host, thành viên đã được duyệt, hoặc đang xin chờ duyệt (pending)
  let myActiveMeetups = $derived(
    meetups.filter((m) => {
      if (!currentUser) return false;
      const isHost = m.hostUid === currentUser.uid || m.host_uid === currentUser.uid;
      const approvedList = m.approvedUids || [];
      const isMember = approvedList.includes(currentUser.uid);
      const pendingList = m.pendingUids || [];
      const isPending = pendingList.includes(currentUser.uid);
      return isHost || isMember || isPending;
    })
  );

  function handleCreateSuccess() {
    selectedLat = null;
    selectedLng = null;
    addressText = '';
    // Switch to active meetups tab to show user their newly created meetup
    activeTab = 'active';
  }
</script>

<section id="create-route" style="padding-bottom: 80px;">
  <!-- Tab selectors -->
  <div class="create-tabs">
    <button class="create-tab-btn {activeTab === 'new' ? 'active' : ''}" onclick={() => activeTab = 'new'}>
      ✨ Tạo kèo mới
    </button>
    <button class="create-tab-btn {activeTab === 'active' ? 'active' : ''}" onclick={() => activeTab = 'active'}>
      🔥 Kèo đang hoạt động ({myActiveMeetups.length})
    </button>
  </div>

  {#if activeTab === 'new'}
    <CreateMeetupForm
      bind:selectedLat
      bind:selectedLng
      bind:addressText
      {userLat}
      {userLng}
      onCreateSuccess={handleCreateSuccess}
    />
  {:else}
    {#if !currentUser}
      <div class="empty-state">
        <p>🔒 Vui lòng đăng nhập ở tab <strong>Hồ sơ</strong> để xem các kèo bạn đang tham gia hoặc chủ trì.</p>
      </div>
    {:else if myActiveMeetups.length === 0}
      <div class="empty-state">
        <p>Bạn chưa chủ trì hay tham gia kèo nào đang hoạt động cả.</p>
        <button class="btn btn-primary" onclick={() => activeTab = 'new'} style="margin-top: 12px;">
          Tạo kèo đầu tiên ngay!
        </button>
      </div>
    {:else}
      <MeetupList
        meetups={myActiveMeetups}
        {userLat}
        {userLng}
        selectedCity="all"
        selectedDistance="all"
        isTrackingGPS={false}
        gpsError={false}
      />
    {/if}
  {/if}
</section>

<style>
  .create-tabs {
    display: flex;
    gap: 12px;
    margin-bottom: 24px;
  }
  .create-tab-btn {
    flex: 1;
    padding: 12px 8px;
    font-size: 1.05rem;
    font-weight: 700;
    background-color: #fffdfb;
    border: 3px solid #1e1e24;
    box-shadow: 4px 4px 0px #1e1e24;
    border-radius: 8px;
    cursor: pointer;
    font-family: 'Quicksand', sans-serif;
    transition: all 0.1s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
  }
  .create-tab-btn.active {
    background-color: var(--pastel-yellow, #ffe869);
    box-shadow: 1px 1px 0px #1e1e24;
    transform: translate(3px, 3px);
  }
  .create-tab-btn:hover:not(.active) {
    background-color: var(--pastel-blue, #a4f0fd);
    transform: translate(-1px, -1px);
    box-shadow: 5px 5px 0px #1e1e24;
  }
  .create-tab-btn:active:not(.active) {
    transform: translate(2px, 2px);
    box-shadow: 2px 2px 0px #1e1e24;
  }
  .empty-state {
    padding: 40px 20px;
    text-align: center;
    background-color: #fffdfb;
    border: 3px solid #1e1e24;
    box-shadow: 5px 5px 0px #1e1e24;
    border-radius: 12px;
    font-size: 1.1rem;
    font-weight: 600;
  }
</style>
