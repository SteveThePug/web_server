/**
 * WebSocket-backed chat, used by components/util/Chat.vue.
 *
 * Owns one socket at a time, auto-reconnects with exponential backoff (1s
 * doubling to a 30s ceiling, reset on a successful open) and refuses to reconnect
 * after an intentional disconnect().
 */

import { defineStore } from "pinia";
import { ref, computed } from "vue";
import axios from "axios";

/**
 * Same-origin ws:// or wss:// URL for the chat socket, derived from the page's
 * own protocol and host. Nothing is configurable: nginx (prod) and the Vite
 * proxy (dev) both serve /api/ws from the same origin as the page, and the
 * auth cookies only travel if the origin matches.
 */
function getWebSocketURL() {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/api/ws`;
}

export const useMessagesStore = defineStore("messages", () => {
  const socket = ref(null);
  const messages = ref([]);
  const isConnected = ref(false);
  const lastError = ref(null);
  const INITIAL_RECONNECT_DELAY_MS = 1000;
  const MAX_RECONNECT_DELAY_MS = 30000;

  let intentionalClose = false;
  let reconnectDelay = INITIAL_RECONNECT_DELAY_MS;
  let reconnectTimer = null;

  const messagesCount = computed(() => messages.value.length);

  /** Open the socket (no-op if one already exists) and arm auto-reconnect. */
  function connect() {
    if (socket.value) return;
    intentionalClose = false;
    clearTimeout(reconnectTimer);

    // Every handler checks it still belongs to the current socket. A socket
    // closed by disconnect() can fire onclose after connect() has already
    // opened a replacement; without this guard it would null the new socket
    // and schedule a second reconnect, leaving two live sockets that each
    // push every broadcast (duplicate messages).
    const ws = new WebSocket(getWebSocketURL());
    socket.value = ws;

    ws.onopen = () => {
      if (socket.value !== ws) return;
      isConnected.value = true;
      lastError.value = null;
      reconnectDelay = INITIAL_RECONNECT_DELAY_MS;
    };

    ws.onmessage = (event) => {
      if (socket.value !== ws) return;
      try {
        const data = JSON.parse(event.data);
        // Sent once per connection. On a reconnect it repeats messages we
        // already have.
        if (data.action === "history") {
          // Replace in one go so keyed rows (and their images) are patched in
          // place rather than unmounted and recreated.
          messages.value = data.messages || [];
          return;
        }
        if (data.action === "delete") {
          messages.value = messages.value.filter((m) => m.id !== data.id);
          return;
        }
        // Dedup by id: the server can re-broadcast a message we already hold
        // (e.g. our own send echoed back, or an overlap with `history`).
        if (data.id && messages.value.some((m) => m.id === data.id)) return;
        messages.value.push(data);
      } catch {
        // Non-JSON frame: show it verbatim rather than dropping it.
        messages.value.push({ text: event.data });
      }
    };

    ws.onerror = (error) => {
      if (socket.value !== ws) return;
      lastError.value = error;
    };

    ws.onclose = () => {
      if (socket.value !== ws) return;
      isConnected.value = false;
      socket.value = null;
      if (!intentionalClose) {
        reconnectTimer = setTimeout(() => {
          connect();
        }, reconnectDelay);
        reconnectDelay = Math.min(reconnectDelay * 2, MAX_RECONNECT_DELAY_MS);
      }
    };
  }

  /**
   * Close the socket and suppress auto-reconnect. Clearing `socket.value`
   * *before* calling close() is what makes the onclose guard above see a stale
   * socket, so a later connect() is not fought over by the dying one.
   */
  function disconnect() {
    intentionalClose = true;
    clearTimeout(reconnectTimer);
    if (!socket.value) return;
    const ws = socket.value;
    socket.value = null;
    isConnected.value = false;
    ws.close();
  }

  /** Send a chat line. Silently drops it if the socket is not open. */
  function sendMessage(text, isPrivate = false) {
    if (!socket.value || !isConnected.value) return;
    socket.value.send(JSON.stringify({ text, private: isPrivate }));
  }

  /** Ask the server to delete a message; the removal arrives back as a
   *  broadcast "delete" frame, so local state is not touched here. */
  function deleteMessage(id) {
    if (!socket.value || !isConnected.value) return;
    socket.value.send(JSON.stringify({ action: "delete", id }));
  }

  function clearMessages() {
    messages.value = [];
  }

  /**
   * POST a file to the REST upload endpoint, then send its returned URL over
   * the socket as a normal message. Two steps because the socket carries JSON
   * text only. Errors surface via `lastError`.
   */
  async function uploadAndSendFile(file, isPrivate = false) {
    try {
      const formData = new FormData();
      formData.append("file", file);
      const res = await axios.post("/api/messages/upload", formData);
      const { url } = res.data;
      if (!socket.value || !isConnected.value) return;
      socket.value.send(
        JSON.stringify({ text: "", fileUrl: url, private: isPrivate }),
      );
    } catch (err) {
      lastError.value = err;
    }
  }

  return {
    messages,
    isConnected,
    lastError,

    messagesCount,

    connect,
    disconnect,
    sendMessage,
    deleteMessage,
    clearMessages,
    uploadAndSendFile,
  };
});
