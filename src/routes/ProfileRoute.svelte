<script lang="ts">
  import Auth from '../lib/Auth.svelte';

  interface Props {
    apiStatus: 'online' | 'offline' | 'connecting';
    apiMessage: string;
    apiCode: number | null;
    apiBase: string;
    isChecking: boolean;
    onCheckHealth: () => void;
  }

  let { apiStatus, apiMessage, apiCode, apiBase, isChecking, onCheckHealth }: Props = $props();
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
          <span class="console-val">/api/health</span>
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

      <div class="console-controls">
        <div class="led-indicator">
          {#if apiStatus === 'online'}
            <span class="led-dot led-online"></span><span>ONLINE</span>
          {:else if apiStatus === 'offline'}
            <span class="led-dot led-offline"></span><span>OFFLINE</span>
          {:else}
            <span class="led-dot led-connecting"></span><span>PINGING</span>
          {/if}
        </div>
        <div class="console-buttons">
          <button
            class="btn btn-success"
            style="padding: 8px 16px; font-size: 0.9rem;"
            onclick={onCheckHealth}
            disabled={isChecking}
          >
            {isChecking ? 'Pinging...' : 'PING API'}
          </button>
        </div>
      </div>
    </section>
  </div>
</section>
