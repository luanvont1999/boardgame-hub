<script lang="ts">
  import CreateMeetupForm from '../lib/CreateMeetupForm.svelte';
  import { navigateToTab } from '../lib/router.svelte';

  interface Props {
    selectedLat: number | null;
    selectedLng: number | null;
    addressText: string;
    userLat: number | null;
    userLng: number | null;
  }

  let {
    selectedLat = $bindable(null),
    selectedLng = $bindable(null),
    addressText = $bindable(''),
    userLat,
    userLng,
  }: Props = $props();

  function handleCreateSuccess() {
    selectedLat = null;
    selectedLng = null;
    addressText = '';
    navigateToTab('find');
  }
</script>

<section id="create-route" style="padding-bottom: 40px;">
  <CreateMeetupForm
    bind:selectedLat
    bind:selectedLng
    bind:addressText
    {userLat}
    {userLng}
    onCreateSuccess={handleCreateSuccess}
  />
</section>
