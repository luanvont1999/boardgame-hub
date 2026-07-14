<script lang="ts">
  import { onMount, onDestroy, tick } from 'svelte';
  import { 
    collection, 
    addDoc, 
    query, 
    orderBy, 
    onSnapshot, 
    serverTimestamp,
    type Unsubscribe
  } from 'firebase/firestore';
  import { onAuthStateChanged, type User } from 'firebase/auth';
  import { db, auth } from './firebase';

  interface Meetup {
    id: string;
    title: string;
    game: string;
    host_name?: string;
    hostName?: string;
    host_uid?: string;
    hostUid?: string;
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
    isOpen: boolean;
    onClose: () => void;
  }

  let { meetup, isOpen, onClose }: Props = $props();

  let messages = $state<ChatMessage[]>([]);
  let newMessageText = $state<string>('');
  let isSending = $state<boolean>(false);
  let currentUser = $state<User | null>(auth.currentUser);
  let messagesContainer = $state<HTMLElement | null>(null);
  let unsubscribeStore: Unsubscribe | null = null;
  let unsubscribeAuth: Unsubscribe | null = null;

  onMount(() => {
    unsubscribeAuth = onAuthStateChanged(auth, (user) => {
      currentUser = user;
    });
  });

  onDestroy(() => {
    if (unsubscribeAuth) unsubscribeAuth();
    if (unsubscribeStore) unsubscribeStore();
  });

  // Re-subscribe when meetup changes or modal opens
  $effect(() => {
    if (isOpen && meetup?.id) {
      subscribeToMessages(meetup.id);
    } else {
      if (unsubscribeStore) {
        unsubscribeStore();
        unsubscribeStore = null;
      }
      messages = [];
    }
  });

  function subscribeToMessages(meetupId: string) {
    if (unsubscribeStore) unsubscribeStore();

    const messagesRef = collection(db, 'meetups', meetupId, 'messages');
    const q = query(messagesRef, orderBy('createdAt', 'asc'));

    unsubscribeStore = onSnapshot(q, (snapshot) => {
      messages = snapshot.docs.map((doc) => ({
        id: doc.id,
        ...doc.data()
      })) as ChatMessage[];

      // Scroll to bottom on new messages
      tick().then(() => {
        if (messagesContainer) {
          messagesContainer.scrollTop = messagesContainer.scrollHeight;
        }
      });
    }, (error) => {
      console.error('[Firestore Chat] Subscription error:', error);
    });
  }

  async function handleSendMessage(e?: Event) {
    if (e) e.preventDefault();
    if (!newMessageText.trim() || !meetup?.id || !currentUser || isSending) return;

    const text = newMessageText.trim();
    newMessageText = '';
    isSending = true;

    try {
      const senderName = currentUser.displayName || currentUser.email || 'Thành viên';
      await addDoc(collection(db, 'meetups', meetup.id, 'messages'), {
        text,
        senderUid: currentUser.uid,
        senderName,
        createdAt: serverTimestamp()
      });
    } catch (err) {
      console.error('[Firestore Chat] Send error:', err);
    } finally {
      isSending = false;
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

  function formatMessageTime(timestamp: any): string {
    if (!timestamp) return 'Vừa xong';
    try {
      const date = timestamp.toDate ? timestamp.toDate() : new Date(timestamp);
      return date.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit' });
    } catch {
      return 'Vừa xong';
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
    <div class="cartoon-card cartoon-modal-content chat-modal-box">
      <!-- Modal Header -->
      <div class="modal-header chat-modal-header">
        <div class="chat-header-title">
          <span class="chat-header-badge">💬 Chat Kèo</span>
          <h3>{meetup.title}</h3>
          <span class="chat-game-tag">🎲 {meetup.game}</span>
        </div>
        <button type="button" class="btn btn-close-modal" onclick={onClose}>✕</button>
      </div>

      <!-- Chat Messages Container -->
      <div class="modal-body chat-modal-body" bind:this={messagesContainer}>
        {#if messages.length === 0}
          <div class="chat-empty-state">
            <span class="empty-icon">💬</span>
            <h4>Chưa có tin nhắn nào!</h4>
            <p>Hãy gửi tin nhắn đầu tiên để hẹn giờ và trao đổi cùng hội nhóm nhé.</p>
          </div>
        {:else}
          <div class="messages-list">
            {#each messages as msg (msg.id)}
              {@const isMine = currentUser && msg.senderUid === currentUser.uid}
              {@const isHost = (meetup.host_uid || meetup.hostUid) === msg.senderUid}
              
              <div class="message-row {isMine ? 'my-message-row' : 'other-message-row'}">
                <div class="message-bubble {isMine ? 'my-bubble' : 'other-bubble'}">
                  <div class="message-author">
                    <span class="author-name">{msg.senderName}</span>
                    {#if isHost}
                      <span class="host-badge">👑 Host</span>
                    {/if}
                  </div>
                  <div class="message-text">{msg.text}</div>
                  <div class="message-time">{formatMessageTime(msg.createdAt)}</div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Footer Input Area -->
      <div class="modal-footer chat-modal-footer">
        {#if currentUser}
          <form class="chat-input-form" onsubmit={handleSendMessage}>
            <input 
              type="text" 
              class="cartoon-chat-input" 
              placeholder="Nhập tin nhắn của bạn..." 
              bind:value={newMessageText}
              disabled={isSending}
            />
            <button 
              type="submit" 
              class="btn btn-primary send-btn" 
              disabled={!newMessageText.trim() || isSending}
            >
              {isSending ? '...' : 'Gửi 🚀'}
            </button>
          </form>
        {:else}
          <div class="login-chat-prompt">
            <span>🔒 Bạn cần đăng nhập (tab Hồ sơ) để gửi tin nhắn chat realtime.</span>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .chat-modal-box {
    max-width: 600px !important;
    height: 82vh;
  }

  .chat-modal-header {
    background-color: var(--pastel-cyan) !important;
  }

  .chat-header-title {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
    text-align: left;
  }

  .chat-header-badge {
    font-size: 0.8rem;
    font-weight: 800;
    padding: 3px 10px;
    background-color: #fff;
    border: var(--border-width-sm) solid var(--color-border);
    border-radius: 100px;
    box-shadow: 1.5px 1.5px 0 var(--color-border);
  }

  .chat-header-title h3 {
    font-size: 1.15rem;
    font-weight: 800;
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 260px;
  }

  .chat-game-tag {
    font-size: 0.8rem;
    font-weight: 700;
    color: var(--text-dark);
    background-color: var(--pastel-yellow);
    padding: 2px 8px;
    border: var(--border-width-sm) solid var(--color-border);
    border-radius: var(--radius-sm);
  }

  .chat-modal-body {
    padding: 16px;
    background-color: var(--bg-secondary);
    display: flex;
    flex-direction: column;
  }

  .chat-empty-state {
    margin: auto;
    text-align: center;
    padding: 32px 16px;
    color: var(--text-muted);
  }

  .empty-icon {
    font-size: 3rem;
    display: block;
    margin-bottom: 12px;
  }

  .chat-empty-state h4 {
    font-size: 1.2rem;
    margin-bottom: 6px;
    color: var(--text-dark);
  }

  .chat-empty-state p {
    font-size: 0.9rem;
    font-weight: 600;
    max-width: 320px;
    margin: 0 auto;
  }

  .messages-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    width: 100%;
  }

  .message-row {
    display: flex;
    width: 100%;
  }

  .my-message-row {
    justify-content: flex-end;
  }

  .other-message-row {
    justify-content: flex-start;
  }

  .message-bubble {
    max-width: 80%;
    padding: 10px 14px;
    border-radius: 18px;
    border: var(--border-width) solid var(--color-border);
    box-shadow: 3px 3px 0 var(--color-border);
    text-align: left;
    display: flex;
    flex-direction: column;
    gap: 4px;
    animation: bubble-pop 0.15s cubic-bezier(0.175, 0.885, 0.32, 1.25);
  }

  @keyframes bubble-pop {
    from { transform: scale(0.85); opacity: 0; }
    to { transform: scale(1); opacity: 1; }
  }

  .my-bubble {
    background-color: var(--pastel-purple);
    border-bottom-right-radius: 4px;
  }

  .other-bubble {
    background-color: #fff;
    border-bottom-left-radius: 4px;
  }

  .message-author {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .author-name {
    font-size: 0.8rem;
    font-weight: 800;
    color: var(--text-dark);
  }

  .host-badge {
    font-size: 0.7rem;
    font-weight: 800;
    background-color: var(--pastel-yellow);
    border: 1.5px solid var(--color-border);
    padding: 1px 6px;
    border-radius: 100px;
  }

  .message-text {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-dark);
    word-break: break-word;
    line-height: 1.4;
  }

  .message-time {
    font-size: 0.7rem;
    font-weight: 700;
    color: var(--text-muted);
    align-self: flex-end;
  }

  /* Chat Input Footer */
  .chat-modal-footer {
    padding: 12px 16px !important;
  }

  .chat-input-form {
    display: flex;
    width: 100%;
    gap: 10px;
  }

  .cartoon-chat-input {
    flex: 1;
    padding: 10px 14px;
    font-family: var(--font-family);
    font-size: 0.95rem;
    font-weight: 600;
    border-radius: var(--radius-md);
    border: var(--border-width) solid var(--color-border);
    background-color: #fff;
    outline: none;
    box-shadow: 2px 2px 0 var(--color-border);
    transition: all 0.15s ease;
  }

  .cartoon-chat-input:focus {
    box-shadow: 3.5px 3.5px 0 var(--color-border);
    border-color: var(--color-border);
  }

  .send-btn {
    padding: 10px 18px !important;
    font-size: 0.95rem !important;
    white-space: nowrap;
  }

  .login-chat-prompt {
    width: 100%;
    text-align: center;
    font-size: 0.85rem;
    font-weight: 700;
    color: var(--text-muted);
    padding: 6px 0;
  }
</style>
