<script lang="ts">
  import Auth from "../lib/Auth.svelte";
  import { getDocs, collection } from "firebase/firestore";
  import { db } from "../lib/firebase";

  interface Props {
    apiStatus: "online" | "offline" | "connecting";
    apiMessage: string;
    apiCode: number | null;
    apiBase: string;
    isChecking: boolean;
    onCheckHealth: () => void;
    addToast: (msg: string, type: "success" | "error" | "info") => void;
  }

  let {
    apiStatus,
    apiMessage,
    apiCode,
    apiBase,
    isChecking,
    onCheckHealth,
    addToast,
  }: Props = $props();

  let isBroadcasting = $state(false);
  let debugLogs = $state<string[]>([]);

  async function handleBroadcastTest() {
    if (isBroadcasting) return;
    isBroadcasting = true;
    debugLogs = [];
    debugLogs.push("> [FCM] Đang quét Firestore lấy danh sách token...");
    addToast("🔄 Đang quét thiết bị và gửi thông báo thử nghiệm...", "info");

    try {
      const querySnapshot = await getDocs(collection(db, "users"));
      const tokens: string[] = [];

      querySnapshot.forEach((doc) => {
        const data = doc.data();
        if (data && data.fcmToken) {
          tokens.push(data.fcmToken);
          debugLogs.push(
            `+ Tìm thấy token của UID: ${doc.id.substring(0, 6)}...`,
          );
        }
      });

      debugLogs.push(`> [FCM] Tổng số thiết bị tìm thấy: ${tokens.length}`);
      if (tokens.length === 0) {
        debugLogs.push("! [ERROR] Không tìm thấy thiết bị nào đăng ký token.");
        addToast("Không tìm thấy thiết bị nào đăng ký token!", "error");
        isBroadcasting = false;
        return;
      }

      debugLogs.push("> [FCM] Đang gửi yêu cầu đẩy đến Backend Go...");
      const API_BASE = apiBase || "";
      const res = await fetch(`${API_BASE}/api/send-notification`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          fcmTokens: tokens,
          title: "📣 Thông báo thử nghiệm!",
          body: "Hệ thống thông báo đẩy Boardgame Luna đang hoạt động bình thường trên thiết bị di động của bạn!",
          clickAction: "/?route=profile",
        }),
      });

      debugLogs.push(
        `> [FCM] HTTP Status từ Backend: ${res.status} ${res.statusText}`,
      );

      const text = await res.text();
      debugLogs.push(
        `> [FCM] Nội dung phản hồi thô: ${text.substring(0, 120)}${text.length > 120 ? "..." : ""}`,
      );

      let data;
      try {
        data = JSON.parse(text);
      } catch (parseErr: any) {
        debugLogs.push(`! [PARSE ERROR] Phản hồi không phải JSON hợp lệ.`);
        addToast("Lỗi định dạng phản hồi (Không phải JSON)!", "error");
        isBroadcasting = false;
        return;
      }

      if (data.success) {
        debugLogs.push(`* [SUCCESS] ${data.message}`);
        addToast(data.message, "success");
      } else {
        debugLogs.push(`! [WARNING] ${data.warning || "Gửi thất bại"}`);
        addToast(data.warning || "Gửi broadcast thất bại!", "error");
      }

      if (data.errors && data.errors.length > 0) {
        debugLogs.push("> Chi tiết lỗi gửi FCM:");
        data.errors.forEach((errStr: string) => {
          debugLogs.push(`  - Lỗi: ${errStr}`);
        });
      }
    } catch (err: any) {
      debugLogs.push(`! [SYSTEM ERROR]: ${err.message}`);
      addToast("Lỗi hệ thống: " + err.message, "error");
    } finally {
      isBroadcasting = false;
    }
  }
</script>

<section id="profile-route" style="padding-bottom: 40px;">
  <h2 class="section-title">Hồ Sơ Của Bạn</h2>
  <Auth />

  <!-- Dev Tools – API Status Console -->
  <div style="margin-top: 40px;">
    <h3
      style="font-size: 1.1rem; color: var(--text-muted); font-weight: 700; margin-bottom: 12px;"
    >
      CÔNG CỤ PHÁT TRIỂN (API STATUS)
    </h3>
    <section class="retro-console" id="status" style="margin-top: 0;">
      <div class="console-screen">
        <div class="console-text-row">
          <span class="console-label">SERVER ADDR:</span>
          <span class="console-val">{apiBase || "Auto (Relative)"}</span>
        </div>
        <div class="console-text-row">
          <span class="console-label">ENDPOINT:</span>
          <span class="console-val">/api/health & /api/send-notification</span>
        </div>
        <div class="console-text-row">
          <span class="console-label">STATUS CODE:</span>
          <span class="console-val">{apiCode !== null ? apiCode : "---"}</span>
        </div>
        <div class="console-text-row">
          <span class="console-label">RESPONSE:</span>
          <span
            class="console-val {apiStatus === 'online'
              ? 'online'
              : apiStatus === 'offline'
                ? 'offline'
                : 'connecting'}"
          >
            {apiMessage}
          </span>
        </div>

        <!-- Live Console Logger -->
        {#if debugLogs.length > 0}
          <div
            style="border-top: 1.5px dashed #1e1e24; margin-top: 12px; padding-top: 10px; font-family: monospace; font-size: 0.8rem; color: #a4f0fd; text-align: left; max-height: 180px; overflow-y: auto; white-space: pre-wrap;"
          >
            {#each debugLogs as log}
              <div style="margin-bottom: 4px; line-height: 1.3;">{log}</div>
            {/each}
          </div>
        {/if}
      </div>

      <div
        class="console-controls"
        style="display: flex; flex-direction: column; gap: 15px; align-items: stretch;"
      >
        <div class="led-indicator">
          {#if apiStatus === "online"}
            <span class="led-dot led-online"></span><span>ONLINE</span>
          {:else if apiStatus === "offline"}
            <span class="led-dot led-offline"></span><span>OFFLINE</span>
          {:else}
            <span class="led-dot led-connecting"></span><span>PINGING</span>
          {/if}
        </div>
        <div
          class="console-buttons"
          style="display: flex; flex-wrap: wrap; gap: 10px;"
        >
          <button
            class="btn btn-success"
            style="padding: 8px 16px; font-size: 0.9rem; flex: 1; min-width: 100px;"
            onclick={onCheckHealth}
            disabled={isChecking}
          >
            {isChecking ? "Pinging..." : "PING API"}
          </button>

          <button
            class="btn btn-primary"
            style="padding: 8px 16px; font-size: 0.9rem; flex: 1; min-width: 160px; background-color: var(--pastel-yellow, #ffe869) !important; color: #1e1e24 !important;"
            onclick={handleBroadcastTest}
            disabled={isBroadcasting}
          >
            {isBroadcasting ? "Broadcasting..." : "BROADCAST TEST 📣"}
          </button>
        </div>
      </div>
    </section>
  </div>
</section>
