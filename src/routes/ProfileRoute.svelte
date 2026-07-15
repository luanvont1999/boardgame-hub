<script lang="ts">
  import { onMount } from "svelte";
  import Auth from "../lib/Auth.svelte";
  import { doc, getDoc, setDoc } from "firebase/firestore";
  import { db, auth } from "../lib/firebase";
  import { onAuthStateChanged, updateProfile, type User } from "firebase/auth";

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
    { id: "strategy", label: "Strategy (Chiến thuật) 🧠" },
    { id: "party", label: "Party Game (Vui nhộn) 🎉" },
    { id: "family", label: "Family (Gia đình) 🏠" },
    { id: "coop", label: "Co-op (Hợp tác) 🤝" },
    { id: "bluffing", label: "Bluffing (Ẩn vai) 🎭" },
    { id: "rpg", label: "RPG (Nhập vai) ⚔️" },
    { id: "economic", label: "Economic (Kinh tế) 📊" },
    { id: "engine", label: "Engine Building (Xây dựng) ⚙️" }
  ];

  let currentUser = $state<User | null>(null);
  let isSaving = $state(false);
  let isLoading = $state(true);

  // Form states
  let displayName = $state("");
  let bio = $state("");
  let favoriteCategories = $state<string[]>([]);

  onMount(() => {
    const unsubscribe = onAuthStateChanged(auth, async (user) => {
      currentUser = user;
      if (user) {
        isLoading = true;
        try {
          const docRef = doc(db, "users", user.uid);
          const docSnap = await getDoc(docRef);
          if (docSnap.exists()) {
            const data = docSnap.data();
            displayName = data.displayName || user.displayName || "";
            bio = data.bio || "";
            favoriteCategories = data.favoriteCategories || [];
          } else {
            displayName = user.displayName || "";
            bio = "";
            favoriteCategories = [];
          }
        } catch (err) {
          console.error("Lỗi tải thông tin profile từ Firestore:", err);
        } finally {
          isLoading = false;
        }
      } else {
        displayName = "";
        bio = "";
        favoriteCategories = [];
        isLoading = false;
      }
    });

    return unsubscribe;
  });

  function toggleCategory(catId: string) {
    if (favoriteCategories.includes(catId)) {
      favoriteCategories = favoriteCategories.filter(id => id !== catId);
    } else {
      favoriteCategories = [...favoriteCategories, catId];
    }
  }

  async function handleSaveProfile(e: Event) {
    e.preventDefault();
    if (!currentUser) return;
    isSaving = true;

    try {
      // 1. Cập nhật profile Firebase Auth
      await updateProfile(currentUser, {
        displayName: displayName
      });

      // 2. Cập nhật thông tin Firestore
      const userRef = doc(db, "users", currentUser.uid);
      await setDoc(userRef, {
        displayName,
        bio,
        favoriteCategories,
        updatedAt: new Date()
      }, { merge: true });

      addToast("🎉 Cập nhật thông tin cá nhân thành công!", "success");
    } catch (err: any) {
      console.error("Lỗi cập nhật profile:", err);
      addToast("Không thể cập nhật hồ sơ: " + err.message, "error");
    } finally {
      isSaving = false;
    }
  }
</script>

<section id="profile-route" style="padding-bottom: 60px;">
  <h2 class="section-title">Hồ Sơ Của Bạn</h2>
  <Auth />

  <!-- PWA Install Section (Neo-brutalist Style) -->
  {#if deferredPrompt || (isIOS && !isStandalone)}
    <div class="cartoon-card install-pwa-card" style="margin-top: 24px; padding: 20px; background-color: #fffefb; text-align: left;">
      <h3 style="font-size: 1.25rem; font-weight: 700; margin-bottom: 8px; color: var(--text-dark);">📱 Cài đặt Ứng dụng PWA</h3>
      <p style="font-size: 0.9rem; font-weight: 500; color: var(--text-muted); margin-bottom: 16px; line-height: 1.4;">
        Cài đặt Boardgame Luna về màn hình chính điện thoại của bạn để trải nghiệm tốc độ mượt mà hơn và nhận thông báo đẩy tức thì như một ứng dụng di động thực thụ!
      </p>
      <button 
        type="button" 
        class="btn btn-warning install-pwa-btn" 
        style="padding: 10px 20px; font-size: 0.95rem; font-weight: 800; background-color: var(--pastel-yellow, #ffe869) !important; color: #1e1e24 !important; border: 3px solid #1e1e24; box-shadow: 4px 4px 0px #1e1e24;"
        onclick={onTriggerInstall}
      >
        Cài đặt ứng dụng ngay 🚀
      </button>
    </div>
  {/if}

  <!-- Profile Edit Section -->
  <div style="margin-top: 40px;">
    <h3 style="font-size: 1.1rem; color: var(--text-muted); font-weight: 700; margin-bottom: 16px; text-transform: uppercase;">
      ⚙️ Thiết lập hồ sơ cá nhân
    </h3>

    {#if isLoading}
      <div class="cartoon-card" style="padding: 24px; text-align: center;">
        <div class="spinner" style="margin: 0 auto 12px auto;"></div>
        <p style="font-weight: 500; color: var(--text-muted);">Đang tải dữ liệu hồ sơ...</p>
      </div>
    {:else if currentUser}
      <form onsubmit={handleSaveProfile} class="cartoon-card profile-edit-form" style="padding: 24px; background-color: #fffefb; text-align: left; display: flex; flex-direction: column; gap: 20px;">
        
        <div class="form-group" style="display: flex; flex-direction: column; gap: 8px;">
          <label for="displayName" style="font-weight: 700; color: var(--text-dark); font-size: 0.95rem;">Biệt danh hiển thị:</label>
          <input 
            type="text" 
            id="displayName" 
            placeholder="Tên của bạn chơi..." 
            bind:value={displayName}
            required
            style="padding: 12px; font-size: 0.95rem; font-weight: 600; border: 3px solid #1e1e24; border-radius: 8px; background-color: #ffffff; box-shadow: 3px 3px 0px #1e1e24; outline: none;"
          />
        </div>

        <div class="form-group" style="display: flex; flex-direction: column; gap: 8px;">
          <label for="bio" style="font-weight: 700; color: var(--text-dark); font-size: 0.95rem;">Lời tự giới thiệu:</label>
          <textarea 
            id="bio" 
            placeholder="Kinh nghiệm chơi, câu nói ưa thích hoặc nhóm boardgame đang hoạt động..." 
            bind:value={bio}
            rows="3"
            style="padding: 12px; font-size: 0.95rem; font-weight: 600; border: 3px solid #1e1e24; border-radius: 8px; background-color: #ffffff; box-shadow: 3px 3px 0px #1e1e24; outline: none; resize: vertical;"
          ></textarea>
        </div>

        <div class="form-group" style="display: flex; flex-direction: column; gap: 10px;">
          <span style="font-weight: 700; color: var(--text-dark); font-size: 0.95rem;">Thể loại boardgame yêu thích:</span>
          <div class="categories-chips-grid" style="display: flex; flex-wrap: wrap; gap: 10px; margin-top: 4px;">
            {#each BOARDGAME_CATEGORIES as cat}
              {@const isSelected = favoriteCategories.includes(cat.id)}
              <button
                type="button"
                class="category-chip-btn {isSelected ? 'selected' : ''}"
                onclick={() => toggleCategory(cat.id)}
                style="padding: 8px 14px; font-size: 0.85rem; font-weight: 700; border: 3px solid #1e1e24; border-radius: 100px; cursor: pointer; transition: transform 0.1s ease, box-shadow 0.1s ease; outline: none;
                       background-color: {isSelected ? 'var(--pastel-yellow, #ffe869)' : '#ffffff'};
                       box-shadow: {isSelected ? '2px 2px 0px #1e1e24' : '3px 3px 0px #1e1e24'};
                       transform: {isSelected ? 'translate(1px, 1px)' : 'none'};"
              >
                {cat.label}
              </button>
            {/each}
          </div>
        </div>

        <button 
          type="submit" 
          class="btn btn-primary" 
          disabled={isSaving}
          style="align-self: flex-start; padding: 12px 24px; font-size: 0.95rem; font-weight: 800; border: 3px solid #1e1e24; border-radius: 8px; background-color: #9ee3b2 !important; color: #1e1e24 !important; box-shadow: 4px 4px 0px #1e1e24; cursor: pointer;"
        >
          {isSaving ? "Đang lưu..." : "Lưu thông tin hồ sơ 💾"}
        </button>

      </form>
    {:else}
      <div class="cartoon-card" style="padding: 32px; background-color: #fffefb; text-align: center;">
        <span style="font-size: 2.5rem; display: block; margin-bottom: 12px;">🔒</span>
        <h4 style="font-size: 1.15rem; font-weight: 700; margin-bottom: 8px; color: var(--text-dark);">Hồ Sơ Chưa Được Kích Hoạt</h4>
        <p style="font-size: 0.9rem; font-weight: 500; color: var(--text-muted); max-width: 320px; margin: 0 auto;">
          Vui lòng đăng nhập hoặc tạo tài khoản mới ở trên để bắt đầu tùy chỉnh thông tin cá nhân của bạn chơi!
        </p>
      </div>
    {/if}
  </div>
</section>

<style>
  /* Local spinner */
  .spinner {
    width: 32px;
    height: 32px;
    border: 4px solid rgba(30, 30, 37, 0.1);
    border-top: 4px solid #1e1e24;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  /* Chip hover effects */
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
</style>
