<script lang="ts">
  import { onMount } from "svelte";
  import { auth } from "../lib/firebase";
  import { onAuthStateChanged, type User } from "firebase/auth";
  import CreateMeetupForm from "../lib/CreateMeetupForm.svelte";
  import { navigateToTab } from "../lib/router.svelte";
  import Icon from "../lib/Icon.svelte";

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
    navigateToTab("my-meetups");
  }
</script>

<section id="create-route">
  <h2 class="section-title">Lên Kèo Chơi Mới</h2>
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
    <div class="cartoon-card locked-card">
      <div class="locked-icon">
        <Icon name="lock" size={40} />
      </div>
      <h4 class="locked-title">Cần Đăng Nhập Tài Khoản</h4>
      <p class="locked-description">
        Bạn cần đăng nhập tài khoản trước khi có thể tự thiết lập và lên kèo
        chơi boardgame mới nhé!
      </p>
      <a href="#/profile" class="btn btn-primary login-link-btn">
        <Icon name="key" size={16} />
        <span>Đi tới trang Đăng nhập</span>
      </a>
    </div>
  {/if}
</section>

<style>
  #create-route {
    padding-bottom: 80px;
  }

  .locked-card {
    padding: 40px;
    background-color: #fffefb;
    text-align: center;
    margin-top: 20px;
  }

  .locked-icon {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-bottom: 16px;
    color: var(--text-dark);
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
    display: inline-flex;
    align-items: center;
    gap: 8px;
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
