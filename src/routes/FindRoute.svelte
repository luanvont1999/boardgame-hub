<script lang="ts">
  import MeetupList from '../lib/MeetupList.svelte';
  import { navigate } from '../lib/router';

  interface Props {
    meetups: any[];
    filteredMeetups: any[];
    selectedCity: 'all' | 'HCM' | 'HN';
    selectedDistance: 'all' | '5' | '10';
    userLat: number | null;
    userLng: number | null;
    isTrackingGPS: boolean;
    gpsError: boolean;
  }

  let {
    meetups,
    filteredMeetups,
    selectedCity = $bindable('all'),
    selectedDistance = $bindable('all'),
    userLat,
    userLng,
    isTrackingGPS,
    gpsError,
  }: Props = $props();
</script>

<section id="find-route" style="padding-bottom: 40px;">
  <!-- Hero Card -->
  <div
    class="cartoon-card hero-discover-card"
  >
    <span class="badge-tag" style="background-color: #fff; margin-bottom: 12px;">Bản Đồ Hội Nhóm 🗺️</span>
    <h2 style="font-size: 1.8rem; margin-bottom: 8px;">Khám phá các nhóm chơi quanh bạn</h2>
    <p style="font-weight: 600; color: var(--text-dark); max-width: 600px; margin: 0 auto 24px; line-height: 1.5; font-size: 0.95rem;">
      Bạn muốn tìm các bàn chơi đang hoạt động trực quan địa lý? Hãy mở bản đồ để quét các quân Meeple màu sắc!
    </p>

    <div class="hero-action-row">
      <button
        class="btn btn-primary"
        style="font-size: 1.05rem; padding: 12px 28px;"
        onclick={() => navigate({ name: 'map', mode: 'discover' })}
      >
        🗺️ Mở Bản Đồ Hội Nhóm
      </button>
      <button
        class="btn btn-secondary"
        style="font-size: 1.05rem; padding: 12px 20px;"
        onclick={() => navigate({ name: 'filter' })}
      >
        🔍 Lọc Kèo
        {#if selectedCity !== 'all' || selectedDistance !== 'all'}
          <span class="filter-active-dot"></span>
        {/if}
      </button>
    </div>
  </div>

  <!-- Meetup Cards List -->
  <MeetupList
    meetups={filteredMeetups}
    {userLat}
    {userLng}
    bind:selectedCity
    bind:selectedDistance
    {isTrackingGPS}
    {gpsError}
  />
</section>

<style>
  .hero-discover-card {
    background-color: var(--pastel-cyan);
    padding: 32px 24px;
    text-align: center;
    margin-bottom: 32px;
    position: relative;
  }

  .hero-action-row {
    display: flex;
    justify-content: center;
    gap: 16px;
    flex-wrap: wrap;
  }

  .filter-active-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    background-color: #ef4444;
    border-radius: 50%;
    margin-left: 6px;
    border: 1.5px solid var(--color-border);
    vertical-align: middle;
  }
</style>
