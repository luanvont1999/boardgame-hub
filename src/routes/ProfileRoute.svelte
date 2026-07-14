<script lang="ts">
  import Auth from '../lib/Auth.svelte';
  import { broadcastPushNotifications } from '../lib/notificationService';

  interface Props {
    apiStatus: 'online' | 'offline' | 'connecting';
    apiMessage: string;
    apiCode: number | null;
    apiBase: string;
    isChecking: boolean;
    onCheckHealth: () => void;
    addToast: (msg: string, type: 'success' | 'error' | 'info') => void;
  }

  let { apiStatus, apiMessage, apiCode, apiBase, isChecking, onCheckHealth, addToast }: Props = $props();

  let isBroadcasting = $state(false);

  async function handleBroadcastTest() {
    if (isBroadcasting) return;
    isBroadcasting = true;
    addToast("🔄 Đang quét thiết bị và gửi thông báo thử nghiệm...", "info");
    
    const result = await broadcastPushNotifications(
      "📣 Thông báo thử nghiệm!",
      "Hệ thống thông báo đẩy Boardgame Luna đang hoạt động bình thường trên thiết bị di động của bạn!"
    );

    if (result.success) {
      addToast(result.message, "success");
    } else {
      addToast(result.message, "error");
    }
    isBroadcasting = false;
  }
</script>

<section id="profile-route" style="padding-bottom: 40px;">
  <h2 class="section-title">Hồ Sơ Của Bạn</h2>
  <Auth />

  <!-- Dev Tools – API Status Console -->
  <div style="margin-top: 40px;">
    <h3 style="font-size: 1.1rem; color: var(--text-muted); font-weight: 700; margin-bottom: 12px;">
      CÔNG CỤ PHÁT TRIỂN (API STATUS)
    </h3>
    <section class="retro-console" id="status" style="margin-top: 0;">
      <div class="console-screen">
        <div class="console-text-row">
          <span class="console-label">SERVER ADDR:</span>
          <span class="console-val">{apiBase || 'Auto (Relative)'}</span>
        </div>
        <div class="console-text-row">
          <span class="console-label">ENDPOINT:</span>
          <span class="console-val">/api/health & /api/send-notification</span>
        </div>
        <div class="console-text-row">
          <span class="console-label">STATUS CODE:</span>
          <span class="console-val">{apiCode !== null ? apiCode : '---'}</span>
        </div>
        <div class="console-text-row">
          <span class="console-label">RESPONSE:</span>
          <span class="console-val {apiStatus === 'online' ? 'online' : apiStatus === 'offline' ? 'offline' : 'connecting'}">
            {apiMessage}
          </span>
        </div>
      </div>

      <div class="console-controls" style="display: flex; flex-direction: column; gap: 15px; align-items: stretch;">
        <div class="led-indicator">
          {#if apiStatus === 'online'}
            <span class="led-dot led-online"></span><span>ONLINE</span>
          {:else if apiStatus === 'offline'}
            <span class="led-dot led-offline"></span><span>OFFLINE</span>
          {:else}
            <span class="led-dot led-connecting"></span><span>PINGING</span>
          {/if}
        </div>
        <div class="console-buttons" style="display: flex; flex-wrap: wrap; gap: 10px;">
          <button
            class="btn btn-success"
            style="padding: 8px 16px; font-size: 0.9rem; flex: 1; min-width: 100px;"
            onclick={onCheckHealth}
            disabled={isChecking}
          >
            {isChecking ? 'Pinging...' : 'PING API'}
          </button>
          
          <button
            class="btn btn-primary"
            style="padding: 8px 16px; font-size: 0.9rem; flex: 1; min-width: 160px; background-color: var(--pastel-yellow, #ffe869) !important; color: #1e1e24 !important;"
            onclick={handleBroadcastTest}
            disabled={isBroadcasting}
          >
            {isBroadcasting ? 'Broadcasting...' : 'BROADCAST TEST 📣'}
          </button>
        </div>
      </div>
    </section>
  </div>
</section>
