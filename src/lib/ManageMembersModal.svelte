<script lang="ts">
  import { onDestroy } from 'svelte';
  import { 
    subscribeToMeetupRequests, 
    approveMember, 
    rejectMember, 
    kickOrLeaveMember, 
    type MeetupRequest 
  } from './meetupService';
  import Icon from './Icon.svelte';

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
    isOpen: boolean;
    onClose: () => void;
  }

  let { meetup, isOpen, onClose }: Props = $props();

  let requests = $state<MeetupRequest[]>([]);
  let isProcessing = $state<boolean>(false);
  let unsubscribeReqs: (() => void) | null = null;

  onDestroy(() => {
    if (unsubscribeReqs) unsubscribeReqs();
  });

  $effect(() => {
    if (isOpen && meetup?.id) {
      if (unsubscribeReqs) unsubscribeReqs();
      unsubscribeReqs = subscribeToMeetupRequests(meetup.id, (list) => {
        requests = list;
      });
    } else {
      if (unsubscribeReqs) {
        unsubscribeReqs();
        unsubscribeReqs = null;
      }
      requests = [];
    }
  });

  let pendingRequests = $derived(requests.filter(r => r.status === 'pending'));
  let approvedRequests = $derived(requests.filter(r => r.status === 'approved'));

  async function handleApprove(playerUid: string) {
    if (!meetup?.id || isProcessing) return;
    isProcessing = true;
    try {
      await approveMember(meetup.id, playerUid);
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
    } catch (err) {
      console.error('Kick error:', err);
    } finally {
      isProcessing = false;
    }
  }

  function handleBackdropClick(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (target.classList.contains('cartoon-modal-backdrop')) {
      onClose();
    }
  }

  function handleBackdropKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape' || e.key === 'Enter') {
      onClose();
    }
  }
</script>

{#if isOpen && meetup}
  <div 
    class="cartoon-modal-backdrop" 
    onclick={handleBackdropClick} 
    onkeydown={handleBackdropKeyDown}
    role="button"
    tabindex="0"
  >
    <div class="cartoon-card cartoon-modal-content manage-modal-box">
      <div class="modal-header manage-modal-header">
        <div class="manage-header-title">
          <span><Icon name="users" size={16} class="inline-icon" /> Quản Lý Thành Viên Kèo</span>
          <h3>{meetup.title}</h3>
        </div>
        <button type="button" class="btn btn-close-modal" onclick={onClose}><Icon name="x" size={16} /></button>
      </div>

      <div class="modal-body manage-modal-body">
        <!-- Host info banner -->
        <div class="host-summary-pill">
          <span><Icon name="crown" size={15} class="inline-icon" /> Host kèo: <strong>{meetup.hostName || meetup.host_name || 'Ẩn danh'}</strong></span>
          <span class="count-tag">
            <Icon name="users" size={13} class="inline-icon" /> {1 + approvedRequests.length} / {meetup.playersNeeded || meetup.players_needed || 4}
          </span>
        </div>

        <!-- Section 1: Yêu cầu tham gia mới (Pending) -->
        <div class="members-section">
          <h4 class="section-heading">
            <Icon name="clock" size={16} class="inline-icon" /> Yêu cầu mới cần xét duyệt ({pendingRequests.length}):
          </h4>

          {#if pendingRequests.length === 0}
            <p class="empty-list-hint">Chưa có ai gửi yêu cầu tham gia mới.</p>
          {:else}
            <div class="requests-list">
              {#each pendingRequests as req (req.uid)}
                <div class="member-card pending-card">
                  <div class="member-info">
                    <span class="member-avatar"><Icon name="user" size={16} /></span>
                    <span class="member-name">{req.name}</span>
                  </div>

                  <div class="member-actions">
                    <button 
                      class="btn btn-success btn-sm" 
                      onclick={() => handleApprove(req.uid)}
                      disabled={isProcessing}
                    >
                      <Icon name="check" size={12} class="inline-icon" /> Duyệt
                    </button>
                    <button 
                      class="btn btn-secondary btn-sm" 
                      onclick={() => handleReject(req.uid)}
                      disabled={isProcessing}
                    >
                      <Icon name="x" size={12} class="inline-icon" /> Từ chối
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Section 2: Thành viên đã tham gia (Approved) -->
        <div class="members-section" style="margin-top: 24px;">
          <h4 class="section-heading">
            <Icon name="check-circle" size={16} class="inline-icon" /> Thành viên trong kèo ({approvedRequests.length}):
          </h4>

          {#if approvedRequests.length === 0}
            <p class="empty-list-hint">Chưa có người chơi nào khác tham gia.</p>
          {:else}
            <div class="requests-list">
              {#each approvedRequests as req (req.uid)}
                <div class="member-card approved-card">
                  <div class="member-info">
                    <span class="member-avatar"><Icon name="user" size={16} /></span>
                    <span class="member-name">{req.name}</span>
                  </div>

                  <button 
                    class="btn btn-secondary btn-sm btn-kick" 
                    onclick={() => handleKick(req.uid)}
                    disabled={isProcessing}
                  >
                    <Icon name="log-out" size={12} class="inline-icon" /> Đuổi khỏi kèo
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>

      <div class="modal-footer">
        <span class="modal-footer-tip">
          <Icon name="sparkles" size={14} class="inline-icon" /> Host có quyền xét duyệt hoặc đuổi thành viên khỏi kèo bất cứ lúc nào.
        </span>
        <button type="button" class="btn btn-primary" onclick={onClose}>Đóng bảng quản lý</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .manage-modal-box {
    max-width: 580px !important;
  }

  .manage-modal-header {
    background-color: var(--pastel-yellow) !important;
  }

  .manage-header-title {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
  }

  .manage-header-title span {
    font-size: 0.8rem;
    font-weight: 800;
    color: var(--text-dark);
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .manage-header-title h3 {
    font-size: 1.2rem;
    font-weight: 800;
    margin: 0;
  }

  .manage-modal-body {
    display: flex;
    flex-direction: column;
    gap: 16px;
    text-align: left;
  }

  .host-summary-pill {
    display: flex;
    justify-content: space-between;
    align-items: center;
    background-color: var(--pastel-cyan);
    padding: 10px 16px;
    border: var(--border-width) solid var(--color-border);
    border-radius: var(--radius-md);
    box-shadow: 2px 2px 0 var(--color-border);
    font-size: 0.9rem;
  }

  .count-tag {
    font-weight: 800;
    background-color: #fff;
    padding: 2px 8px;
    border: 1.5px solid var(--color-border);
    border-radius: 100px;
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .section-heading {
    font-size: 1rem;
    font-weight: 800;
    margin-bottom: 10px;
    color: var(--text-dark);
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .empty-list-hint {
    font-size: 0.88rem;
    font-weight: 600;
    color: var(--text-muted);
    font-style: italic;
  }

  .requests-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .member-card {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 14px;
    background-color: #fff;
    border: var(--border-width) solid var(--color-border);
    border-radius: var(--radius-md);
    box-shadow: 2.5px 2.5px 0 var(--color-border);
  }

  .member-info {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .member-avatar {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .member-name {
    font-size: 0.95rem;
    font-weight: 800;
  }

  .member-actions {
    display: flex;
    gap: 6px;
  }

  .btn-sm {
    padding: 6px 12px !important;
    font-size: 0.82rem !important;
    border-radius: var(--radius-sm) !important;
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .btn-kick {
    background-color: var(--pastel-pink) !important;
  }
</style>
