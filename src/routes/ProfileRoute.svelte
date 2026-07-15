<script lang="ts">
  import { onMount } from "svelte";
  import Auth from "../lib/Auth.svelte";
  import Icon from "../lib/Icon.svelte";
  import { db, auth } from "../lib/firebase";
  import { onAuthStateChanged, updateProfile, type User } from "firebase/auth";
  import { userProfileState } from "../lib/userProfile.svelte";

  interface Props {
    addToast: (msg: string, type: "success" | "error" | "info") => void;
    deferredPrompt: any;
    isIOS: boolean;
    isStandalone: boolean;
    onTriggerInstall: () => void;
  }

  let {
    addToast,
    deferredPrompt,
    isIOS,
    isStandalone,
    onTriggerInstall,
  }: Props = $props();

  const BOARDGAME_CATEGORIES = [
    { id: "strategy", label: "Strategy (Chiến thuật)", icon: "strategy" },
    { id: "party", label: "Party Game (Vui nhộn)", icon: "party" },
    { id: "family", label: "Family (Gia đình)", icon: "family" },
    { id: "coop", label: "Co-op (Hợp tác)", icon: "coop" },
    { id: "bluffing", label: "Bluffing (Ẩn vai)", icon: "bluffing" },
    { id: "rpg", label: "RPG (Nhập vai)", icon: "rpg" },
    { id: "economic", label: "Economic (Kinh tế)", icon: "economic" },
    { id: "engine", label: "Engine Building (Xây dựng)", icon: "building-cog" },
  ];

  let currentUser = $state<User | null>(null);
  let isSaving = $state(false);

  // Form states
  let displayName = $state("");
  let bio = $state("");
  let favoriteCategories = $state<string[]>([]);

  // Bind to global state loading indicator
  let isLoading = $derived(userProfileState.isLoading);

  // Synchronize form inputs when global cache updates or user changes
  $effect(() => {
    if (userProfileState.profile.loaded) {
      displayName = userProfileState.profile.displayName;
      bio = userProfileState.profile.bio;
      favoriteCategories = userProfileState.profile.favoriteCategories;
    }
  });

  onMount(() => {
    const unsubscribe = onAuthStateChanged(auth, (user) => {
      currentUser = user;
    });
    return unsubscribe;
  });

  function toggleCategory(catId: string) {
    if (favoriteCategories.includes(catId)) {
      favoriteCategories = favoriteCategories.filter((id) => id !== catId);
    } else {
      favoriteCategories = [...favoriteCategories, catId];
    }
  }

  async function handleSaveProfile(e: Event) {
    e.preventDefault();
    if (!currentUser) return;
    isSaving = true;

    try {
      await updateProfile(currentUser, {
        displayName: displayName,
      });

      await userProfileState.updateProfile({
        displayName,
        bio,
        favoriteCategories,
      });

      addToast("Cập nhật thông tin cá nhân thành công!", "success");
    } catch (err: any) {
      console.error("Lỗi cập nhật profile:", err);
      addToast("Không thể cập nhật hồ sơ: " + err.message, "error");
    } finally {
      isSaving = false;
    }
  }
</script>

<section id="profile-route">
  <h2 class="section-title">Hồ Sơ Của Bạn</h2>
  <Auth />

  <!-- PWA Install Section -->
  {#if deferredPrompt || (isIOS && !isStandalone)}
    <div class="cartoon-card install-pwa-card">
      <h3 class="pwa-title">
        <Icon name="smartphone" size={20} class="inline-icon" /> Cài đặt Ứng dụng PWA
      </h3>
      <p class="pwa-description">
        Cài đặt Boardgame Luna về màn hình chính điện thoại của bạn để trải
        nghiệm tốc độ mượt mà hơn và nhận thông báo đẩy tức thì như một ứng dụng
        di động thực thụ!
      </p>
      <button
        type="button"
        class="btn install-pwa-btn"
        onclick={onTriggerInstall}
      >
        <Icon name="rocket" size={18} class="inline-icon" /> Cài đặt ứng dụng ngay
      </button>
    </div>
  {/if}

  <!-- Profile Edit Section -->
  <div class="profile-edit-section">
    <h3 class="profile-section-heading">
      <Icon name="settings" size={18} class="inline-icon" /> Thiết lập hồ sơ cá nhân
    </h3>

    {#if isLoading}
      <div class="cartoon-card profile-edit-form skeleton-profile-form">
        <!-- Skeleton: Display Name -->
        <div class="form-group">
          <div class="skeleton skeleton-line short"></div>
          <div class="skeleton skeleton-input"></div>
        </div>
        <!-- Skeleton: Bio -->
        <div class="form-group">
          <div class="skeleton skeleton-line short"></div>
          <div class="skeleton skeleton-textarea"></div>
        </div>
        <!-- Skeleton: Categories -->
        <div class="form-group">
          <div class="skeleton skeleton-line medium"></div>
          <div class="categories-chips-grid">
            {#each [1, 2, 3, 4, 5] as _}
              <div
                class="skeleton skeleton-badge"
                style="width: 120px; height: 32px;"
              ></div>
            {/each}
          </div>
        </div>
        <!-- Skeleton: Save Button -->
        <div
          class="skeleton skeleton-btn"
          style="width: 200px; height: 46px;"
        ></div>
      </div>
    {:else if currentUser}
      <form onsubmit={handleSaveProfile} class="cartoon-card profile-edit-form">
        <div class="form-group">
          <label for="displayName">Tên hiển thị:</label>
          <input
            type="text"
            id="displayName"
            placeholder="Tên của bạn chơi..."
            bind:value={displayName}
            required
          />
        </div>

        <div class="form-group">
          <label for="bio">Giới thiệu:</label>
          <textarea
            id="bio"
            placeholder="Kinh nghiệm chơi, câu nói ưa thích hoặc nhóm boardgame đang hoạt động..."
            bind:value={bio}
            rows="3"
          ></textarea>
        </div>

        <div class="form-group">
          <span class="form-label">Thể loại boardgame yêu thích:</span>
          <div class="categories-chips-grid">
            {#each BOARDGAME_CATEGORIES as cat}
              {@const isSelected = favoriteCategories.includes(cat.id)}
              <button
                type="button"
                class="category-chip-btn {isSelected ? 'selected' : ''}"
                onclick={() => toggleCategory(cat.id)}
              >
                <Icon name={cat.icon} size={15} class="chip-icon" />
                <span>{cat.label}</span>
              </button>
            {/each}
          </div>
        </div>

        <button
          type="submit"
          class="btn save-profile-btn {isSaving ? 'btn-loading' : ''}"
          disabled={isSaving}
        >
          <Icon name="save" size={18} class="inline-icon" />
          <span>{isSaving ? "Đang lưu..." : "Lưu thông tin hồ sơ"}</span>
        </button>
      </form>
    {:else}
      <div class="cartoon-card locked-card">
        <div class="locked-icon">
          <Icon name="lock" size={40} />
        </div>
        <h4 class="locked-title">Hồ Sơ Chưa Được Kích Hoạt</h4>
        <p class="locked-description">
          Vui lòng đăng nhập hoặc tạo tài khoản mới ở trên để bắt đầu tùy chỉnh
          thông tin cá nhân của bạn chơi!
        </p>
      </div>
    {/if}
  </div>
</section>

<style>
  #profile-route {
    padding-bottom: 60px;
  }

  /* PWA Install Card */
  .install-pwa-card {
    padding: 20px;
    background-color: #fffefb;
    text-align: left;
  }

  .pwa-title {
    font-size: 1.25rem;
    font-weight: 700;
    margin-bottom: 8px;
    color: var(--text-dark);
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .pwa-description {
    font-size: 0.9rem;
    font-weight: 500;
    color: var(--text-muted);
    margin-bottom: 16px;
    line-height: 1.4;
  }

  .install-pwa-btn {
    padding: 10px 20px;
    font-size: 0.95rem;
    font-weight: 800;
    background-color: var(--pastel-yellow, #ffe869);
    color: #1e1e24;
    border: 3px solid #1e1e24;
    box-shadow: 4px 4px 0px #1e1e24;
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  /* Profile Edit Section */
  .profile-edit-section {
    margin-top: 20px;
  }

  .profile-section-heading {
    font-size: 1.1rem;
    color: var(--text-muted);
    font-weight: 700;
    margin-bottom: 16px;
    text-transform: uppercase;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  /* Skeleton Profile Form */
  .skeleton-profile-form .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  /* Profile Edit Form */
  .profile-edit-form {
    padding: 24px;
    background-color: #fffefb;
    text-align: left;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .profile-edit-form .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .profile-edit-form label,
  .form-label {
    font-weight: 700;
    color: var(--text-dark);
    font-size: 0.95rem;
  }

  .profile-edit-form input,
  .profile-edit-form textarea {
    padding: 12px;
    font-size: 0.95rem;
    font-weight: 600;
    border: 3px solid #1e1e24;
    border-radius: 8px;
    background-color: #ffffff;
    box-shadow: 3px 3px 0px #1e1e24;
    outline: none;
    font-family: var(--font-family);
  }

  .profile-edit-form textarea {
    resize: vertical;
  }

  /* Categories Chips */
  .categories-chips-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 4px;
  }

  .category-chip-btn {
    padding: 8px 14px;
    font-size: 0.85rem;
    font-weight: 700;
    font-family: var(--font-family);
    border: 3px solid #1e1e24;
    border-radius: 100px;
    cursor: pointer;
    transition:
      transform 0.1s ease,
      box-shadow 0.1s ease;
    outline: none;
    background-color: #ffffff;
    box-shadow: 3px 3px 0px #1e1e24;
    display: inline-flex;
    align-items: center;
  }

  .category-chip-btn.selected {
    background-color: var(--pastel-yellow, #ffe869);
    box-shadow: 2px 2px 0px #1e1e24;
    transform: translate(1px, 1px);
  }

  .category-chip-btn:hover {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px 0px #1e1e24;
  }

  .category-chip-btn.selected:hover {
    transform: translate(1px, 1px);
    box-shadow: 2px 2px 0px #1e1e24;
  }

  .category-chip-btn:active {
    transform: translate(3px, 3px);
    box-shadow: 0px 0px 0px #1e1e24;
  }

  /* Save Button */
  .save-profile-btn {
    align-self: flex-start;
    padding: 12px 24px;
    font-size: 0.95rem;
    font-weight: 800;
    border: 3px solid #1e1e24;
    border-radius: 8px;
    background-color: #9ee3b2;
    color: #1e1e24;
    box-shadow: 4px 4px 0px #1e1e24;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  /* Locked State */
  .locked-card {
    padding: 32px;
    background-color: #fffefb;
    text-align: center;
  }

  .locked-icon {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-bottom: 12px;
    color: var(--text-dark);
  }

  .locked-title {
    font-size: 1.15rem;
    font-weight: 700;
    margin-bottom: 8px;
    color: var(--text-dark);
  }

  .locked-description {
    font-size: 0.9rem;
    font-weight: 500;
    color: var(--text-muted);
    max-width: 320px;
    margin: 0 auto;
  }
</style>
