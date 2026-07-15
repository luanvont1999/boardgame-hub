<script lang="ts">
  import MeetupList from "../lib/MeetupList.svelte";
  import { navigate } from "../lib/router.svelte";

  interface Props {
    meetups: any[];
    filteredMeetups: any[];
    selectedCity: "all" | "HCM" | "HN";
    selectedDistance: "all" | "5" | "10";
    userLat: number | null;
    userLng: number | null;
    isTrackingGPS: boolean;
    gpsError: boolean;
    isLoading?: boolean;
  }

  let {
    meetups,
    filteredMeetups,
    selectedCity = $bindable("all"),
    selectedDistance = $bindable("all"),
    userLat,
    userLng,
    isTrackingGPS,
    gpsError,
    isLoading = false,
  }: Props = $props();
</script>

<section id="find-route" style="padding-bottom: 40px;">
  <h2 class="section-title">Tìm kèo xung quanh</h2>
  <!-- Meetup Cards List -->
  <MeetupList
    meetups={filteredMeetups}
    {userLat}
    {userLng}
    bind:selectedCity
    bind:selectedDistance
    {isTrackingGPS}
    {gpsError}
    {isLoading}
  />
</section>

<style>
</style>
