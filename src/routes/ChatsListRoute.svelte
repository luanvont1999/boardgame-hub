<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import {
    collection,
    query,
    orderBy,
    limit,
    onSnapshot,
    type Unsubscribe,
  } from "firebase/firestore";
  import { onAuthStateChanged, type User } from "firebase/auth";
  import { db, auth } from "../lib/firebase";
  import { navigate } from "../lib/router.svelte";
  import Icon from "../lib/Icon.svelte";

  interface Props {
    meetups: any[];
  }

  let { meetups = [] }: Props = $props();

  let currentUser = $state<User | null>(auth.currentUser);
  let latestMessages = $state<
    Record<string, { text: string; senderName: string; createdAt: any }>
  >({});
  let unsubscribes: Unsubscribe[] = [];
  let authUnsubscribe: Unsubscribe | null = null;

  onMount(() => {
    authUnsubscribe = onAuthStateChanged(auth, (u) => {
      currentUser = u;
    });
  });

  onDestroy(() => {
    if (authUnsubscribe) authUnsubscribe();
    unsubscribes.forEach((unsub) => unsub());
  });

  // Filter meetups where currentUser is Host or Approved Member
  let myMeetups = $derived.by(() => {
    const user = currentUser;
    if (!user) return [];
    return meetups.filter(
      (m) =>
        m.hostUid === user.uid ||
        m.host_uid === user.uid ||
        (Array.isArray(m.approvedUids) && m.approvedUids.includes(user.uid)),
    );
  });

  // Subscribe to latest message for each joined meetup
  $effect(() => {
    unsubscribes.forEach((unsub) => unsub());
    unsubscribes = [];

    myMeetups.forEach((meetup) => {
      const messagesRef = collection(db, "meetups", meetup.id, "messages");
      const q = query(messagesRef, orderBy("createdAt", "desc"), limit(1));

      const unsub = onSnapshot(
        q,
        (snapshot) => {
          if (!snapshot.empty) {
            const data = snapshot.docs[0].data();
            latestMessages[meetup.id] = {
              text: data.text || "",
              senderName: data.senderName || "Thành viên",
              createdAt: data.createdAt,
            };
          }
        },
        (err) =>
          console.error(
            `[ChatsList] Error fetching message for ${meetup.id}:`,
            err,
          ),
      );
      unsubscribes.push(unsub);
    });
  });

  // Sort joined chatboxes by latest message timestamp descending (Mới nhất lên đầu)
  let sortedChats = $derived.by(() => {
    return [...myMeetups].sort((a, b) => {
      const msgA = latestMessages[a.id];
      const msgB = latestMessages[b.id];

      const timeA = msgA?.createdAt?.seconds
        ? msgA.createdAt.seconds * 1000
        : a.createdAt?.seconds
          ? a.createdAt.seconds * 1000
          : 0;

      const timeB = msgB?.createdAt?.seconds
        ? msgB.createdAt.seconds * 1000
        : b.createdAt?.seconds
          ? b.createdAt.seconds * 1000
          : 0;

      return timeB - timeA;
    });
  });

  function formatTime(timestamp: any): string {
    if (!timestamp) return "";
    try {
      const date = timestamp.seconds
        ? new Date(timestamp.seconds * 1000)
        : new Date(timestamp);
      const now = new Date();
      const isToday = date.toDateString() === now.toDateString();
      const hours = String(date.getHours()).padStart(2, "0");
      const minutes = String(date.getMinutes()).padStart(2, "0");

      if (isToday) return `${hours}:${minutes}`;

      const day = String(date.getDate()).padStart(2, "0");
      const month = String(date.getMonth() + 1).padStart(2, "0");
      return `${day}/${month}`;
    } catch {
      return "";
    }
  }

  function openChat(meetupId: string) {
    navigate({ name: "chat", meetupId });
  }
</script>

<div class="chats-list-route">
  <!-- Route Header -->
  <div class="cartoon-card route-header">
    <h2>
      <Icon name="message-square" size={24} class="inline-icon" /> Trò Chuyện Kèo Chơi
    </h2>
    <p class="sub-title">
      Tất cả các phòng chat của hội nhóm bạn đang tham gia (sắp xếp từ mới nhất)
    </p>
  </div>

  {#if !currentUser}
    <div class="cartoon-card empty-card">
      <Icon name="user" size={32} />
      <h3>Chưa Đăng Nhập</h3>
      <p>Vui lòng đăng nhập để xem danh sách phòng chat của các kèo chơi.</p>
    </div>
  {:else if sortedChats.length === 0}
    <div class="cartoon-card empty-card">
      <Icon name="message-square" size={36} />
      <h3>Chưa Có Phòng Chat Nào</h3>
      <p>
        Bạn chưa tham gia kèo chơi nào. Hãy ra trang <strong>Tìm Kèo</strong> hoặc <strong>Lên Kèo</strong> mới để tham gia trò chuyện!
      </p>
    </div>
  {:else}
    <div class="chats-list">
      {#each sortedChats as meetup (meetup.id)}
        {@const latest = latestMessages[meetup.id]}
        {@const isHost = meetup.hostUid === currentUser?.uid || meetup.host_uid === currentUser?.uid}

        <button
          type="button"
          class="cartoon-card chat-item-card"
          onclick={() => openChat(meetup.id)}
        >
          <div
            class="chat-avatar"
            style="background-color: {meetup.color || '#a4f0fd'};"
          >
            <svg viewBox="0 0 24 24" width="22" height="22">
              <path
                d="M12 2a2.5 2.5 0 0 0-2.5 2.5c0 1.6 1.2 2.5 2.5 2.5s2.5-.9 2.5-2.5A2.5 2.5 0 0 0 12 2z"
                fill="#1e1e24"
              />
              <path
                d="M17 12V10c0-1.2-1.2-2-2.8-2.4H9.8C8.2 8 7 8.8 7 10v2c0 .4.4.8.8.8h.8v6c0 .4.4.8.8.8h5.2c.4 0 .8-.4.8-.8v-6h.8c.4 0 .8-.4.8-.8z"
                fill="#1e1e24"
              />
            </svg>
          </div>

          <div class="chat-info">
            <div class="chat-header-row">
              <h4 class="chat-title">{meetup.title}</h4>
              {#if latest?.createdAt}
                <span class="chat-time">{formatTime(latest.createdAt)}</span>
              {/if}
            </div>

            <div class="chat-sub-row">
              <span class="game-badge">🎲 {meetup.game}</span>
              {#if isHost}
                <span class="host-chip">👑 Host</span>
              {/if}
            </div>

            {#if latest}
              <p class="chat-snippet">
                <strong>{latest.senderName}:</strong> {latest.text}
              </p>
            {:else}
              <p class="chat-snippet no-msg-text">Chưa có tin nhắn nào trong phòng chat...</p>
            {/if}
          </div>

          <Icon name="arrow-right" size={16} class="arrow-icon" />
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .chats-list-route {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding-bottom: 30px;
    max-width: 768px;
    margin: 0 auto;
  }

  .route-header {
    background-color: var(--pastel-pink, #ffd6e0);
    padding: 16px 20px;
    text-align: left;
  }

  .route-header h2 {
    font-size: 1.25rem;
    font-weight: 800;
    margin: 0 0 4px 0;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .sub-title {
    font-size: 0.82rem;
    font-weight: 600;
    color: var(--text-dark);
    margin: 0;
  }

  .empty-card {
    background-color: #fffefb;
    padding: 32px 20px;
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 10px;
  }

  .empty-card h3 {
    font-size: 1.1rem;
    font-weight: 800;
    margin: 0;
  }

  .empty-card p {
    font-size: 0.88rem;
    color: var(--text-muted);
    margin: 0;
  }

  .chats-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .chat-item-card {
    background-color: #fffefb;
    padding: 14px 16px;
    display: flex;
    align-items: center;
    gap: 14px;
    text-align: left;
    cursor: pointer;
    transition: all 0.12s ease;
    border: var(--border-width) solid var(--color-border);
    box-shadow: 3.5px 3.5px 0 var(--color-border);
    border-radius: var(--radius-lg);
  }

  .chat-item-card:hover {
    transform: translate(-2px, -2px);
    box-shadow: 5px 5px 0 var(--color-border);
    background-color: #ffffff;
  }

  .chat-item-card:active {
    transform: translate(1px, 1px);
    box-shadow: 1.5px 1.5px 0 var(--color-border);
  }

  .chat-avatar {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    border: 2px solid #1e1e24;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 2px 2px 0 #1e1e24;
    flex-shrink: 0;
  }

  .chat-info {
    flex: 1;
    min-width: 0;
  }

  .chat-header-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
  }

  .chat-title {
    font-size: 0.98rem;
    font-weight: 800;
    margin: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--text-dark);
  }

  .chat-time {
    font-size: 0.75rem;
    font-weight: 700;
    color: var(--text-muted);
    white-space: nowrap;
  }

  .chat-sub-row {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 2px;
  }

  .game-badge {
    font-size: 0.75rem;
    font-weight: 700;
    color: var(--text-dark);
    opacity: 0.85;
  }

  .host-chip {
    font-size: 0.7rem;
    font-weight: 800;
    background-color: var(--pastel-yellow, #ffe869);
    padding: 1px 6px;
    border-radius: 100px;
    border: 1px solid #1e1e24;
  }

  .chat-snippet {
    font-size: 0.85rem;
    color: var(--text-dark);
    margin: 4px 0 0 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .no-msg-text {
    color: var(--text-muted);
    font-style: italic;
  }
</style>
