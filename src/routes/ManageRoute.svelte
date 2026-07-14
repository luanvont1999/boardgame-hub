<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { 
    subscribeToMeetupRequests, 
    approveMember, 
    rejectMember, 
    kickOrLeaveMember, 
    type MeetupRequest 
  } from '../lib/meetupService';
  import { goBack } from '../lib/router.svelte';
  import { doc, getDoc } from 'firebase/firestore';
  import { db } from '../lib/firebase';
  import { sendPushNotificationProxy } from '../lib/notificationService';

  async function getUserFcmToken(uid: string): Promise<string | null> {
    try {
      const userDoc = await getDoc(doc(db, 'users', uid));
      if (userDoc.exists()) {
        return userDoc.data()?.fcmToken || null;
      }
    } catch (e) {
      console.error("Lỗi lấy FCM token của user:", e);
    }
    return null;
  }

  interface Meetup {
    id: string;
    title: string;
    game: string;
    hostName?: string;
    host_name?: string;
    hostUid?: string;
    host_uid?: string;
    playersCount?: number;
    players_count?: number;
    playersNeeded?: number;
    players_needed?: number;
  }

  interface Props {
    meetup: Meetup | null;
  }

  let { meetup }: Props = $props();

  let requests = $state<MeetupRequest[]>([]);
  let isProcessing = $state<boolean>(false);
  let unsubscribeReqs: (() => void) | null = null;

  onMount(() => {
    if (meetup?.id) {
      subscribe();
    }
  });

  onDestroy(() => {
    if (unsubscribeReqs) unsubscribeReqs();
  });

  $effect(() => {
    if (meetup?.id) {
      subscribe();
    }
  });

  function subscribe() {
    if (!meetup?.id) return;
    if (unsubscribeReqs) unsubscribeReqs();
    unsubscribeReqs = subscribeToMeetupRequests(meetup.id, (list) => {
      requests = list;
    });
  }

  let pendingRequests = $derived(requests.filter(r => r.status === 'pending'));
  let approvedRequests = $derived(requests.filter(r => r.status === 'approved'));

  async function handleApprove(playerUid: string) {
    if (!meetup?.id || isProcessing) return;
    isProcessing = true;
    try {
      await approveMember(meetup.id, playerUid);
      
      const token = await getUserFcmToken(playerUid);
      if (token) {
        const host = meetup.hostName || meetup.host_name || 'Host';
        sendPushNotificationProxy(
          token,
          "🎉 Yêu cầu đã được duyệt!",
          `Bạn đã được duyệt tham gia kèo "${meetup.title}" chơi game ${meetup.game} của ${host}!`,
          `/?route=manage&meetupId=${meetup.id}`
        );
      }
    } catch (err) {
      console.error('Approve error:', err);
    } finally {
      isProcessing = false;
    }
  }

  async function handleReject(playerUid: string) {
    if (!meetup?.id || isProcessing) return;
    isProcessing = true;
    try {
      await rejectMember(meetup.id, playerUid);

      const token = await getUserFcmToken(playerUid);
      if (token) {
        sendPushNotificationProxy(
          token,
          "✕ Yêu cầu tham gia bị từ chối",
          `Rất tiếc, yêu cầu tham gia kèo "${meetup.title}" của bạn đã bị từ chối.`,
          `/`
        );
      }
    } catch (err) {
      console.error('Reject error:', err);
    } finally {
      isProcessing = false;
    }
  }

  async function handleKick(playerUid: string) {
    if (!meetup?.id || isProcessing) return;
    isProcessing = true;
    try {
      await kickOrLeaveMember(meetup.id, playerUid);

      const token = await getUserFcmToken(playerUid);
      if (token) {
        sendPushNotificationProxy(
          token,
          "✕ Bạn đã bị xóa khỏi kèo",
          `Host đã xóa bạn khỏi danh sách tham gia kèo "${meetup.title}".`,
          `/`
        );
      }
    } catch (err) {
      console.error('Kick error:', err);
    } finally {
      isProcessing = false;
    }
  }
</script>

<div class="fullscreen-route-view manage-route">
  {#if !meetup}
    <div class="cartoon-card no-meetup-card">
      <h3>Chưa chọn kèo chơi để quản lý! 😢</h3>
      <button class="btn btn-primary" onclick={goBack}>Quay lại danh sách kèo</button>
    </div>
  {:else}
    <!-- Top Navigation -->
    <div class="cartoon-card route-top-nav">
      <button type="button" class="btn btn-secondary back-btn" onclick={goBack}>
        ← Quay lại
      </button>
      <div class="nav-title-group">
        <h2>👥 Bảng Quản Lý Thành Viên Kèo</h2>
        <span class="sub-title">Host: {meetup.hostName || meetup.host_name || 'Ẩn danh'} • Kèo: {meetup.title}</span>
      </div>
    </div>

    <!-- Manage Body Card -->
    <div class="cartoon-card manage-body-card">
      <div class="summary-bar">
        <span class="meetup-name-badge">🎲 Game: <strong>{meetup.game}</strong></span>
        <span class="player-count-badge">
          👥 Sĩ số: <strong>{meetup.playersCount || meetup.players_count || 1} / {meetup.playersNeeded || meetup.players_needed || 4} người</strong>
        </span>
      </div>

      <!-- Pending Section -->
      <div class="members-block">
        <h3 class="block-title">⏳ Yêu cầu tham gia mới ({pendingRequests.length}):</h3>

        {#if pendingRequests.length === 0}
          <div class="empty-list-box">
            <span>Chưa có yêu cầu tham gia mới nào đang chờ duyệt.</span>
          </div>
        {:else}
          <div class="members-grid">
            {#each pendingRequests as req (req.uid)}
              <div class="member-item-card pending-item">
                <div class="user-profile-row">
                  <span class="user-icon">👤</span>
                  <span class="user-name">{req.name}</span>
                </div>

                <div class="item-actions">
                  <button 
                    class="btn btn-success action-sm-btn" 
                    onclick={() => handleApprove(req.uid)}
                    disabled={isProcessing}
                  >
                    Duyệt vào kèo ✅
                  </button>
                  <button 
                    class="btn btn-secondary action-sm-btn" 
                    onclick={() => handleReject(req.uid)}
                    disabled={isProcessing}
                  >
                    Từ chối ❌
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Approved Members Section -->
      <div class="members-block" style="margin-top: 32px;">
        <h3 class="block-title">✅ Thành viên đã được duyệt ({approvedRequests.length}):</h3>

        {#if approvedRequests.length === 0}
          <div class="empty-list-box">
            <span>Chưa có thành viên nào khác tham gia kèo này.</span>
          </div>
        {:else}
          <div class="members-grid">
            {#each approvedRequests as req (req.uid)}
              <div class="member-item-card approved-item">
                <div class="user-profile-row">
                  <span class="user-icon">🎲</span>
                  <span class="user-name">{req.name}</span>
                </div>

                <button 
                  class="btn btn-secondary action-sm-btn kick-btn" 
                  onclick={() => handleKick(req.uid)}
                  disabled={isProcessing}
                >
                  Đuổi khỏi kèo 🚪
                </button>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .fullscreen-route-view {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 20px;
    padding-bottom: 40px;
  }

  .no-meetup-card {
    margin: auto;
    text-align: center;
    padding: 40px;
    background-color: #fff;
  }

  .no-meetup-card h3 {
    margin-bottom: 16px;
  }

  .route-top-nav {
    background-color: var(--pastel-yellow);
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 16px 24px;
    border-radius: var(--radius-lg);
    box-shadow: 4px 4px 0 var(--color-border);
    text-align: left;
  }

  .back-btn {
    padding: 8px 16px !important;
    font-size: 0.95rem !important;
    white-space: nowrap;
  }

  .nav-title-group h2 {
    font-size: 1.3rem;
    font-weight: 800;
    margin: 0;
  }

  .sub-title {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--text-dark);
  }

  .manage-body-card {
    background-color: #fffefb;
    padding: 28px 24px;
    text-align: left;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .summary-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background-color: var(--pastel-cyan);
    padding: 12px 20px;
    border: var(--border-width) solid var(--color-border);
    border-radius: var(--radius-md);
    box-shadow: 3px 3px 0 var(--color-border);
    font-size: 0.95rem;
    flex-wrap: wrap;
    gap: 12px;
  }

  .player-count-badge {
    background-color: #fff;
    padding: 4px 12px;
    border: var(--border-width-sm) solid var(--color-border);
    border-radius: 100px;
    font-weight: 800;
  }

  .members-block {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .block-title {
    font-size: 1.1rem;
    font-weight: 800;
    color: var(--text-dark);
  }

  .empty-list-box {
    padding: 16px 20px;
    background-color: var(--bg-secondary);
    border: var(--border-width-sm) dashed var(--color-border);
    border-radius: var(--radius-md);
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--text-muted);
  }

  .members-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }

  .member-item-card {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 14px 16px;
    background-color: #fff;
    border: var(--border-width) solid var(--color-border);
    border-radius: var(--radius-md);
    box-shadow: 3px 3px 0 var(--color-border);
    gap: 12px;
  }

  .user-profile-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .user-icon {
    font-size: 1.3rem;
  }

  .user-name {
    font-size: 0.98rem;
    font-weight: 800;
    color: var(--text-dark);
  }

  .item-actions {
    display: flex;
    gap: 8px;
  }

  .action-sm-btn {
    padding: 8px 14px !important;
    font-size: 0.85rem !important;
  }

  .kick-btn {
    background-color: var(--pastel-pink) !important;
  }
</style>
