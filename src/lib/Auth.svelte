<script lang="ts">
  import { onMount } from "svelte";
  import {
    signInWithEmailAndPassword,
    createUserWithEmailAndPassword,
    signInWithPopup,
    signOut,
    onAuthStateChanged,
    getIdToken,
    type User,
  } from "firebase/auth";
  import { auth, googleProvider } from "./firebase";
  import Icon from "./Icon.svelte";

  // Svelte 5 State Runes
  let user = $state<User | null>(null);
  let idToken = $state<string>("");
  let isLoginMode = $state<boolean>(true);
  let email = $state<string>("");
  let password = $state<string>("");

  let isLoading = $state<boolean>(true);
  let isSubmitting = $state<boolean>(false);
  let errorMessage = $state<string>("");
  let successMessage = $state<string>("");
  let isCopied = $state<boolean>(false);

  // Check if Firebase keys are placeholders
  let isConfigMissing = $state<boolean>(false);

  onMount(() => {
    // Basic verification of API Key config
    if (auth.config?.apiKey === "PLACEHOLDER_API_KEY") {
      isConfigMissing = true;
    }

    const unsubscribe = onAuthStateChanged(auth, async (currentUser) => {
      user = currentUser;
      if (currentUser) {
        try {
          idToken = await getIdToken(currentUser, true);
        } catch (err) {
          console.error("Failed to get ID Token", err);
        }
      } else {
        idToken = "";
      }
      isLoading = false;
    });

    return unsubscribe;
  });

  import { userProfileState } from "./userProfile.svelte";

  async function handleEmailAuth(e: Event) {
    e.preventDefault();
    if (!email || !password) {
      errorMessage = "Vui lòng nhập đầy đủ Email và Mật khẩu!";
      return;
    }

    errorMessage = "";
    successMessage = "";
    isSubmitting = true;

    try {
      if (isLoginMode) {
        const cred = await signInWithEmailAndPassword(auth, email, password);
        if (cred.user) {
          await userProfileState.fetchProfile(cred.user.uid, cred.user.displayName);
        }
        successMessage = "Đăng nhập thành công!";
      } else {
        const cred = await createUserWithEmailAndPassword(auth, email, password);
        if (cred.user) {
          await userProfileState.fetchProfile(cred.user.uid, cred.user.displayName);
        }
        successMessage = "Đăng ký tài khoản thành công!";
      }
    } catch (err: any) {
      console.error(err);
      errorMessage = translateError(err.code || err.message);
    } finally {
      isSubmitting = false;
    }
  }

  async function handleGoogleLogin() {
    errorMessage = "";
    successMessage = "";
    isSubmitting = true;

    try {
      const cred = await signInWithPopup(auth, googleProvider);
      if (cred.user) {
        await userProfileState.fetchProfile(cred.user.uid, cred.user.displayName);
      }
      successMessage = "Đăng nhập Google thành công!";
    } catch (err: any) {
      console.error(err);
      errorMessage = translateError(err.code || err.message);
    } finally {
      isSubmitting = false;
    }
  }

  async function handleSignOut() {
    errorMessage = "";
    successMessage = "";
    try {
      await signOut(auth);
      successMessage = "Đã đăng xuất thành công.";
    } catch (err: any) {
      errorMessage = "Lỗi đăng xuất: " + err.message;
    }
  }

  function copyToken() {
    if (!idToken) return;
    navigator.clipboard.writeText(idToken);
    isCopied = true;
    setTimeout(() => {
      isCopied = false;
    }, 2000);
  }

  function translateError(code: string): string {
    switch (code) {
      case "auth/invalid-credential":
        return "Email hoặc mật khẩu không chính xác!";
      case "auth/email-already-in-use":
        return "Email này đã được sử dụng cho một tài khoản khác!";
      case "auth/weak-password":
        return "Mật khẩu phải dài ít nhất 6 ký tự!";
      case "auth/invalid-email":
        return "Định dạng Email không hợp lệ!";
      case "auth/popup-closed-by-user":
        return "Cửa sổ đăng nhập Google đã bị đóng!";
      default:
        return "Có lỗi xảy ra: " + code;
    }
  }
</script>

<div class="auth-wrapper" id="auth-section">
  {#if isConfigMissing}
    <div class="cartoon-card warning-card">
      <div class="warning-header">
        <Icon name="alert-triangle" size={24} />
        <h3>Firebase Chưa Được Cấu Hình!</h3>
      </div>
      <p>
        Bạn cần cập nhật cấu hình Firebase thật của mình bằng cách tạo file <code
          >.env</code
        >
        trong thư mục <code>frontend/</code> và thêm các khóa sau:
      </p>
      <pre><code
          >VITE_FIREBASE_API_KEY=YOUR_KEY
VITE_FIREBASE_AUTH_DOMAIN=YOUR_DOMAIN
VITE_FIREBASE_PROJECT_ID=YOUR_PROJECT_ID
VITE_FIREBASE_STORAGE_BUCKET=YOUR_BUCKET
VITE_FIREBASE_MESSAGING_SENDER_ID=YOUR_SENDER_ID
VITE_FIREBASE_APP_ID=YOUR_APP_ID</code
        ></pre>
      <p
        style="margin-top: 10px; font-size: 0.85rem; color: var(--text-muted);"
      >
        * Bạn có thể lấy các thông tin này bằng cách tạo Project trên <a
          href="https://console.firebase.google.com/"
          target="_blank"
          style="text-decoration: underline; font-weight: bold;"
          >Firebase Console</a
        >.
      </p>
    </div>
  {/if}

  {#if isLoading}
    <!-- Skeleton Auth Card -->
    <div class="cartoon-card profile-card skeleton-auth-card">
      <div class="profile-header">
        <div class="skeleton profile-avatar-placeholder"></div>
        <div class="skeleton-auth-info">
          <div class="skeleton skeleton-line long"></div>
          <div class="skeleton skeleton-line medium"></div>
        </div>
      </div>
      <div class="skeleton skeleton-btn" style="width: 160px; height: 42px;"></div>
    </div>
  {:else if user}
    <!-- Logged In Profile UI -->
    <div class="cartoon-card profile-card">
      <div class="profile-header">
        {#if user.photoURL}
          <img src={user.photoURL} alt="Avatar" class="profile-avatar" />
        {:else}
          <div class="profile-avatar-placeholder">
            {user.email ? user.email[0].toUpperCase() : "U"}
          </div>
        {/if}
        <div class="profile-info">
          <h3>Chào mừng, {user.displayName || "Thành viên"}!</h3>
          <p class="user-email">{user.email}</p>
        </div>
      </div>

      <!-- Token verification area -->

      <div class="auth-actions">
        <button class="btn btn-secondary" onclick={handleSignOut}>
          Đăng xuất tài khoản
        </button>
      </div>
    </div>
  {:else}
    <!-- Login/Register Form -->
    <div class="cartoon-card auth-card">
      <div class="tabs-header">
        <button
          class="tab-btn {isLoginMode ? 'active' : ''}"
          onclick={() => {
            isLoginMode = true;
            errorMessage = "";
          }}
        >
          Đăng nhập
        </button>
        <button
          class="tab-btn {!isLoginMode ? 'active' : ''}"
          onclick={() => {
            isLoginMode = false;
            errorMessage = "";
          }}
        >
          Đăng ký
        </button>
      </div>

      <form onsubmit={handleEmailAuth} class="auth-form">
        {#if errorMessage}
          <div class="alert-box error-alert">
            {errorMessage}
          </div>
        {/if}
        {#if successMessage}
          <div class="alert-box success-alert">
            {successMessage}
          </div>
        {/if}

        <div class="form-group">
          <label for="email">Địa chỉ Email:</label>
          <input
            type="email"
            id="email"
            placeholder="example@gmail.com"
            bind:value={email}
            disabled={isSubmitting}
            required
          />
        </div>

        <div class="form-group">
          <label for="password">Mật khẩu:</label>
          <input
            type="password"
            id="password"
            placeholder="••••••"
            bind:value={password}
            disabled={isSubmitting}
            required
          />
        </div>

        <button
          type="submit"
          class="btn btn-primary w-full {isSubmitting ? 'btn-loading' : ''}"
          disabled={isSubmitting}
        >
          {#if !isSubmitting}
            <Icon name={isLoginMode ? "key" : "sparkles"} size={18} class="inline-icon" />
          {/if}
          <span>
            {isSubmitting
              ? "Đang xử lý..."
              : isLoginMode
                ? "Đăng nhập ngay"
                : "Đăng ký tài khoản"}
          </span>
        </button>
      </form>

      <div class="divider">
        <span>Hoặc</span>
      </div>

      <button
        class="btn btn-google w-full {isSubmitting ? 'btn-loading' : ''}"
        onclick={handleGoogleLogin}
        disabled={isSubmitting}
      >
        <svg class="google-icon" viewBox="0 0 24 24" width="20" height="20">
          <path
            fill="#4285F4"
            d="M23.745 12.27c0-.7-.06-1.4-.19-2.07H12v3.92h6.69a5.74 5.74 0 0 1-2.48 3.77v3.08h3.99c2.34-2.15 3.69-5.32 3.69-8.7z"
          />
          <path
            fill="#34A853"
            d="M12 24c3.24 0 5.97-1.08 7.96-2.91l-3.99-3.08a7.12 7.12 0 0 1-3.97 1.13c-3.06 0-5.65-2.07-6.58-4.85H1.27v3.18A12 12 0 0 0 12 24z"
          />
          <path
            fill="#FBBC05"
            d="M5.42 14.29a7.15 7.15 0 0 1 0-4.58V6.53H1.27a12 12 0 0 0 0 10.94l4.15-3.18z"
          />
          <path
            fill="#EA4335"
            d="M12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42A11.92 11.92 0 0 0 12 0 12 12 0 0 0 1.27 6.53l4.15 3.18c.93-2.78 3.52-4.85 6.58-4.85z"
          />
        </svg>
        Đăng nhập bằng Google
      </button>
    </div>
  {/if}
</div>

<style>
  .auth-wrapper {
    width: 100%;
    max-width: 520px;
    margin: 0 auto 20px;
  }

  /* Warning config card */
  .warning-card {
    background-color: #ffccd3;
    border-color: #ef4444;
    margin-bottom: 24px;
    text-align: left;
  }

  .warning-header {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;
    color: #b91c1c;
  }

  .warning-card pre {
    background-color: rgba(255, 255, 255, 0.5);
    border: 2px solid #ef4444;
    padding: 10px;
    border-radius: var(--radius-sm);
    font-family: monospace;
    font-size: 0.85rem;
    margin: 10px 0;
    overflow-x: auto;
  }

  /* Profile Card styling */
  .profile-card {
    text-align: left;
    background-color: #fffefb;
  }

  .profile-header {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 24px;
  }

  .profile-avatar {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    border: var(--border-width) solid var(--color-border);
    box-shadow: 2px 2px 0 var(--color-border);
  }

  .profile-avatar-placeholder {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    border: var(--border-width) solid var(--color-border);
    background-color: var(--pastel-cyan);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.8rem;
    font-weight: bold;
    box-shadow: 2px 2px 0 var(--color-border);
  }

  .profile-info h3 {
    font-size: 1.4rem;
    color: var(--text-dark);
  }

  .user-email {
    color: var(--text-muted);
    font-weight: 500;
  }

  /* Skeleton Auth Card */
  .skeleton-auth-card .profile-avatar-placeholder {
    background-color: transparent;
    border-color: #e8e2d6;
  }

  .skeleton-auth-info {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  /* Form & Tabs styling */
  .auth-card {
    padding: 24px;
    background-color: #fffefb;
  }

  .tabs-header {
    display: flex;
    border-bottom: var(--border-width) solid var(--color-border);
    margin-bottom: 24px;
  }

  .tab-btn {
    flex: 1;
    padding: 12px;
    font-size: 1.1rem;
    font-weight: 700;
    background: none;
    border: none;
    cursor: pointer;
    font-family: var(--font-family);
    color: var(--text-muted);
    transition: all 0.15s ease;
  }

  .tab-btn.active {
    color: var(--text-dark);
    background-color: var(--pastel-purple);
    border-top: var(--border-width) solid var(--color-border);
    border-left: var(--border-width) solid var(--color-border);
    border-right: var(--border-width) solid var(--color-border);
    border-radius: var(--radius-sm) var(--radius-sm) 0 0;
    transform: translateY(3px); /* align on the border bottom line */
    box-shadow: 2px -2px 0 var(--color-border);
  }

  .auth-form {
    text-align: left;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin-bottom: 18px;
  }

  .form-group label {
    font-weight: 700;
    font-size: 0.95rem;
  }

  .form-group input {
    padding: 12px;
    border-radius: var(--radius-md);
    border: var(--border-width) solid var(--color-border);
    font-family: var(--font-family);
    font-size: 1rem;
    font-weight: 600;
    box-shadow: 3px 3px 0 var(--color-border);
    outline: none;
    transition: all 0.15s ease;
  }

  .form-group input:focus {
    transform: translate(-1px, -1px);
    box-shadow: 4px 4px 0 var(--color-border);
    background-color: #fffdf5;
  }

  .w-full {
    width: 100%;
  }

  .divider {
    display: flex;
    align-items: center;
    text-align: center;
    margin: 20px 0;
    color: var(--text-muted);
    font-weight: 700;
  }

  .divider::before,
  .divider::after {
    content: "";
    flex: 1;
    border-bottom: var(--border-width-sm) dashed var(--color-border);
  }

  .divider span {
    padding: 0 10px;
  }

  .btn-google {
    background-color: #ffffff;
  }

  .google-icon {
    margin-right: 8px;
  }

  /* Alerts */
  .alert-box {
    padding: 10px 14px;
    border-radius: var(--radius-md);
    border: var(--border-width-sm) solid var(--color-border);
    font-weight: 700;
    font-size: 0.9rem;
    margin-bottom: 16px;
  }

  .error-alert {
    background-color: #ffccd3;
    color: #b91c1c;
  }

  .success-alert {
    background-color: #d1fae5;
    color: #065f46;
  }
</style>

