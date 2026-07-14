<script lang="ts">
  import { onMount, onDestroy, tick } from 'svelte';
  import { 
    collection, addDoc, query, orderBy, onSnapshot, serverTimestamp,
    type Unsubscribe
  } from 'firebase/firestore';
  import { onAuthStateChanged, type User } from 'firebase/auth';
  import { db, auth } from '../lib/firebase';
  import { 
    isApprovedMember, requestToJoin, cancelJoinRequest, 
    subscribeToMeetupRequests, type MeetupRequest 
  } from '../lib/meetupService';
  import { goBack, navigateToTab } from '../lib/router.svelte';

  interface Meetup {
    id: string;
    title: string;
    game: string;
    host_name?: string;
    hostName?: string;
    host_uid?: string;
    hostUid?: string;
    approvedUids?: string[];
    color?: string;
  }

  interface ChatMessage {
    id: string;
    text: string;
    senderUid: string;
    senderName: string;
    createdAt?: any;
  }

  interface Props {
    meetup: Meetup | null;
  }

  let { meetup }: Props = $props();

  let messages = $state<ChatMessage[]>([]);
  let requests = $state<MeetupRequest[]>([]);
  let newMessageText = $state<string>('');
  let isSending = $state<boolean>(false);
  let isRequesting = $state<boolean>(false);
  let currentUser = $state<User | null>(auth.currentUser);
  let messagesContainer = $state<HTMLElement | null>(null);
  let unsubscribeMessages: Unsubscribe | null = null;
  let unsubscribeAuth: (() => void) | null = null;
  let unsubscribeReqs: (() => void) | null = null;

  let hasChatAccess = $derived(isApprovedMember(meetup, currentUser?.uid));

  let myRequest = $derived.by(() => {
    if (!currentUser || !requests.length) return null;
    return requests.find(r => r.uid === currentUser?.uid) || null;
  });

  onMount(() => {
    unsubscribeAuth = onAuthStateChanged(auth, (user) => {
      currentUser = user;
    });
    if (meetup?.id) {
      subscribeMessages(meetup.id);
      unsubscribeReqs = subscribeToMeetupRequests(meetup.id, (list) => {
        requests = list;
      });
    }
  });

  onDestroy(() => {
    if (unsubscribeAuth) unsubscribeAuth();
    if (unsubscribeMessages) unsubscribeMessages();
    if (unsubscribeReqs) unsubscribeReqs();
  });

  function subscribeMessages(meetupId: string) {
    if (unsubscribeMessages) unsubscribeMessages();
    const q = query(collection(db, 'meetups', meetupId, 'messages'), orderBy('createdAt', 'asc'));
    unsubscribeMessages = onSnapshot(q, (snapshot) => {
      messages = snapshot.docs.map(doc => ({ id: doc.id, ...doc.data() })) as ChatMessage[];
      tick().then(() => {
        if (messagesContainer) messagesContainer.scrollTop = messagesContainer.scrollHeight;
      });
    });
  }

  async function handleSend(e?: Event) {
    if (e) e.preventDefault();
    if (!newMessageText.trim() || !meetup?.id || !currentUser || isSending) return;
    const text = newMessageText.trim();
    newMessageText = '';
    isSending = true;
    try {
      await addDoc(collection(db, 'meetups', meetup.id, 'messages'), {
        text,
        senderUid: currentUser.uid,
        senderName: currentUser.displayName || currentUser.email || 'Thành viên',
        createdAt: serverTimestamp()
      });
    } finally {
      isSending = false;
    }
  }

  async function handleSendJoinRequest() {
    if (!meetup?.id || !currentUser || isRequesting) return;
    isRequesting = true;
    try { await requestToJoin(meetup.id, currentUser); }
    finally { isRequesting = false; }
  }

  async function handleCancelRequest() {
    if (!meetup?.id || !currentUser || isRequesting) return;
    isRequesting = true;
    try { await cancelJoinRequest(meetup.id, currentUser.uid); }
    finally { isRequesting = false; }
  }

  function formatTime(ts: any): string {
    if (!ts) return 'Vừa xong';
    try {
      const d = ts.toDate ? ts.toDate() : new Date(ts);
      return d.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' });
    } catch { return 'Vừa xong'; }
  }
</script>

<div class="fullscreen-chat-view">
  {#if !meetup}
    <div class="cartoon-card no-chat-card">
      <h3>Chưa chọn kèo chơi nào! 😢</h3>
      <button class="btn btn-primary" onclick={() => navigateToTab('find')}>Quay lại danh sách kèo</button>
    </div>

  {:else if !hasChatAccess}
    <!-- Top Nav -->
    <div class="chat-top-nav cartoon-card">
      <button class="btn btn-secondary back-btn" onclick={goBack}>← Quay lại</button>
      <div class="chat-nav-details">
        <div class="chat-title-row">
          <h2 class="chat-meetup-title">{meetup.title}</h2>
          <span class="game-badge">🎲 {meetup.game}</span>
        </div>
        <span class="host-info">👑 Host: <strong>{meetup.host_name || meetup.hostName || 'Ẩn danh'}</strong></span>
      </div>
    </div>

    <!-- Access Denied -->
    <div class="cartoon-card access-denied-card">
      <span class="lock-big-icon">🔒</span>
      <h3>Phòng Chat Riêng Tư</h3>
      <p>Chỉ những người chơi đã được Host phê duyệt vào kèo mới có quyền xem và nhắn tin trong phòng chat này.</p>
      <div class="access-action-box">
        {#if !currentUser}
          <p>💡 Hãy đăng nhập ở tab <strong>Hồ sơ</strong> để gửi yêu cầu tham gia kèo.</p>
          <button class="btn btn-primary" onclick={goBack}>Quay lại danh sách kèo</button>
        {:else if myRequest?.status === 'pending'}
          <div class="pending-status-pill">⏳ Yêu cầu của bạn đang chờ Host xét duyệt...</div>
          <div class="action-btn-group">
            <button class="btn btn-secondary" onclick={handleCancelRequest} disabled={isRequesting}>
              {isRequesting ? '...' : 'Hủy yêu cầu ✖'}
            </button>
            <button class="btn btn-primary" onclick={goBack}>Quay lại danh sách</button>
          </div>
        {:else}
          <div class="action-btn-group">
            <button class="btn btn-success" onclick={handleSendJoinRequest} disabled={isRequesting}>
              {isRequesting ? 'Đang gửi...' : 'Gửi yêu cầu tham gia kèo 🤝'}
            </button>
            <button class="btn btn-primary" onclick={goBack}>Quay lại danh sách</button>
          </div>
        {/if}
      </div>
    </div>

  {:else}
    <!-- Top Nav -->
    <div class="chat-top-nav cartoon-card">
      <button class="btn btn-secondary back-btn" onclick={goBack}>← Quay lại</button>
      <div class="chat-nav-details">
        <div class="chat-title-row">
          <h2 class="chat-meetup-title">{meetup.title}</h2>
          <span class="game-badge">🎲 {meetup.game}</span>
        </div>
        <span class="host-info">👑 Host: <strong>{meetup.host_name || meetup.hostName || 'Ẩn danh'}</strong></span>
      </div>
    </div>

    <!-- Messages Feed -->
    <div class="chat-feed-area cartoon-card" bind:this={messagesContainer}>
      {#if messages.length === 0}
        <div class="chat-empty-state">
          <span class="empty-icon">💬</span>
          <h3>Chưa có tin nhắn nào trong kèo này!</h3>
          <p>Hãy mở lời gửi tin nhắn đầu tiên để thảo luận và hẹn giờ chơi cùng mọi người nhé.</p>
        </div>
      {:else}
        <div class="messages-list">
          {#each messages as msg (msg.id)}
            {@const isMine = currentUser && msg.senderUid === currentUser.uid}
            {@const isHostMsg = (meetup.host_uid || meetup.hostUid) === msg.senderUid}
            <div class="message-row {isMine ? 'my-message-row' : 'other-message-row'}">
              <div class="message-bubble {isMine ? 'my-bubble' : 'other-bubble'}">
                <div class="message-author">
                  <span class="author-name">{msg.senderName}</span>
                  {#if isHostMsg}<span class="host-badge">👑 Host</span>{/if}
                </div>
                <div class="message-text">{msg.text}</div>
                <div class="message-time">{formatTime(msg.createdAt)}</div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Input Bar -->
    <div class="chat-bottom-bar cartoon-card">
      {#if currentUser}
        <form class="chat-input-form" onsubmit={handleSend}>
          <input
            type="text"
            class="cartoon-chat-input"
            placeholder="Nhập tin nhắn..."
            bind:value={newMessageText}
            disabled={isSending}
          />
          <button type="submit" class="btn btn-primary send-btn" disabled={!newMessageText.trim() || isSending}>
            {isSending ? '...' : 'Gửi 🚀'}
          </button>
        </form>
      {:else}
        <div class="login-chat-prompt">
          🔒 Bạn cần đăng nhập ở tab <strong>Hồ sơ</strong> để gửi tin nhắn chat.
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .fullscreen-chat-view {
    width: 100%;
    display: flex;
    flex-direction: column;
    height: calc(100vh - 80px);
    gap: 16px;
    margin-bottom: 20px;
  }
  @media (max-width: 768px) {
    .fullscreen-chat-view { height: calc(100vh - 80px); }
  }
  .no-chat-card { margin: auto; text-align: center; padding: 40px; background-color: #fff; }
  .access-denied-card { margin: auto; text-align: center; padding: 40px 24px; background-color: #fffefb; max-width: 480px; display: flex; flex-direction: column; align-items: center; gap: 16px; }
  .lock-big-icon { font-size: 3.5rem; }
  .access-denied-card h3 { font-size: 1.5rem; margin: 0; }
  .access-denied-card p { font-size: 0.95rem; font-weight: 600; color: var(--text-muted); line-height: 1.5; margin: 0; }
  .access-action-box { margin-top: 12px; display: flex; flex-direction: column; align-items: center; gap: 14px; width: 100%; }
  .action-btn-group { display: flex; gap: 12px; width: 100%; justify-content: center; }
  .action-btn-group .btn { flex: 1; }
  .pending-status-pill { padding: 10px 16px; background-color: var(--pastel-yellow); border: var(--border-width) solid var(--color-border); border-radius: var(--radius-md); font-weight: 800; font-size: 0.9rem; box-shadow: 3px 3px 0 var(--color-border); }
  .chat-top-nav { background-color: var(--pastel-cyan); display: flex; align-items: center; gap: 16px; padding: 14px 20px; border-radius: var(--radius-lg); box-shadow: 4px 4px 0 var(--color-border); flex-shrink: 0; }
  .back-btn { padding: 8px 16px !important; font-size: 0.95rem !important; white-space: nowrap; }
  .chat-nav-details { display: flex; flex-direction: column; gap: 2px; text-align: left; flex: 1; min-width: 0; }
  .chat-title-row { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
  .chat-meetup-title { font-size: 1.3rem; font-weight: 800; margin: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: var(--text-dark); }
  .game-badge { font-size: 0.82rem; font-weight: 800; background-color: var(--pastel-yellow); padding: 2px 10px; border: var(--border-width-sm) solid var(--color-border); border-radius: var(--radius-sm); box-shadow: 1.5px 1.5px 0 var(--color-border); }
  .host-info { font-size: 0.85rem; font-weight: 600; color: var(--text-dark); }
  .chat-feed-area { flex: 1; background-color: var(--bg-primary); overflow-y: auto; padding: 20px; display: flex; flex-direction: column; border-radius: var(--radius-lg); box-shadow: 4px 4px 0 var(--color-border); }
  .chat-empty-state { margin: auto; text-align: center; padding: 40px 20px; color: var(--text-muted); }
  .empty-icon { font-size: 3.5rem; display: block; margin-bottom: 12px; }
  .chat-empty-state h3 { font-size: 1.3rem; margin-bottom: 8px; color: var(--text-dark); }
  .chat-empty-state p { font-size: 0.95rem; font-weight: 600; max-width: 360px; margin: 0 auto; }
  .messages-list { display: flex; flex-direction: column; gap: 16px; width: 100%; }
  .message-row { display: flex; width: 100%; }
  .my-message-row { justify-content: flex-end; }
  .other-message-row { justify-content: flex-start; }
  .message-bubble { max-width: 75%; padding: 12px 16px; border-radius: 20px; border: var(--border-width) solid var(--color-border); box-shadow: 4px 4px 0 var(--color-border); text-align: left; display: flex; flex-direction: column; gap: 4px; animation: bubble-pop 0.15s cubic-bezier(0.175, 0.885, 0.32, 1.25); }
  @keyframes bubble-pop { from { transform: scale(0.9); opacity: 0; } to { transform: scale(1); opacity: 1; } }
  .my-bubble { background-color: var(--pastel-purple); border-bottom-right-radius: 4px; }
  .other-bubble { background-color: #fff; border-bottom-left-radius: 4px; }
  .message-author { display: flex; align-items: center; gap: 6px; }
  .author-name { font-size: 0.82rem; font-weight: 800; color: var(--text-dark); }
  .host-badge { font-size: 0.7rem; font-weight: 800; background-color: var(--pastel-yellow); border: 1.5px solid var(--color-border); padding: 1px 6px; border-radius: 100px; }
  .message-text { font-size: 1rem; font-weight: 600; color: var(--text-dark); word-break: break-word; line-height: 1.4; }
  .message-time { font-size: 0.72rem; font-weight: 700; color: var(--text-muted); align-self: flex-end; }
  .chat-bottom-bar { background-color: #fff; padding: 14px 20px; border-radius: var(--radius-lg); box-shadow: 4px 4px 0 var(--color-border); flex-shrink: 0; }
  .chat-input-form { display: flex; width: 100%; gap: 12px; }
  .cartoon-chat-input { flex: 1; padding: 12px 16px; font-family: var(--font-family); font-size: 1rem; font-weight: 600; border-radius: var(--radius-md); border: var(--border-width) solid var(--color-border); background-color: var(--bg-secondary); outline: none; box-shadow: 3px 3px 0 var(--color-border); transition: all 0.15s ease; }
  .cartoon-chat-input:focus { background-color: #fff; box-shadow: 4px 4px 0 var(--color-border); }
  .send-btn { padding: 12px 24px !important; font-size: 1rem !important; white-space: nowrap; }
  .login-chat-prompt { width: 100%; text-align: center; font-size: 0.9rem; font-weight: 700; color: var(--text-muted); padding: 6px 0; }
</style>
