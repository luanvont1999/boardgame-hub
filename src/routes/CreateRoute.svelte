<script lang="ts">
  import { onMount } from "svelte";
  import { auth } from "../lib/firebase";
  import { onAuthStateChanged, type User } from "firebase/auth";
  import CreateMeetupForm from "../lib/CreateMeetupForm.svelte";
  import { navigateToTab } from "../lib/router.svelte";

  interface Props {
    meetups?: any[];
    selectedLat: number | null;
    selectedLng: number | null;
    addressText: string;
    userLat: number | null;
    userLng: number | null;
  }

  let {
    selectedLat = $bindable(null),
    selectedLng = $bindable(null),
    addressText = $bindable(""),
    userLat,
    userLng,
  }: Props = $props();

  let currentUser = $state<User | null>(auth.currentUser);

  onMount(() => {
    const unsub = onAuthStateChanged(auth, (user) => {
      currentUser = user;
    });
    return unsub;
  });

  function handleCreateSuccess() {
    selectedLat = null;
    selectedLng = null;
    addressText = "";
    // Chuyển hướng người dùng sang tab "Các Kèo" sau khi tạo thành công để xem danh sách
    navigateToTab("my-meetups");
  }
</script>

<section id="create-route" style="padding-bottom: 80px;">
  {#if currentUser}
    <CreateMeetupForm
      bind:selectedLat
      bind:selectedLng
      bind:addressText
      {userLat}
      {userLng}
      onCreateSuccess={handleCreateSuccess}
    />
  {:else}
    <!-- Unauthorized warning -->
    <div class="cartoon-card" style="padding: 40px; background-color: #fffefb; text-align: center; margin-top: 20px;">
      <span style="font-size: 3rem; display: block; margin-bottom: 16px;">🔒</span>
      <h4 style="font-size: 1.3rem; font-weight: 700; margin-bottom: 8px; color: var(--text-dark);">Cần Đăng Nhập Tài Khoản</h4>
      <p style="font-size: 0.95rem; font-weight: 500; color: var(--text-muted); max-width: 360px; margin: 0 auto 24px auto; line-height: 1.5;">
        Bạn cần đăng nhập tài khoản trước khi có thể tự thiết lập và lên kèo chơi boardgame mới nhé!
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
</style>
