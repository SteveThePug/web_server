import { defineStore } from "pinia";
import { ref, computed } from "vue";
import axios from "axios";

function getWebSocketURL() {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/api/ws`;
}

export const useMessagesStore = defineStore("messages", () => {
  const socket = ref(null);
  const messages = ref([]);
  const isConnected = ref(false);
  const lastError = ref(null);
  let intentionalClose = false;
  let reconnectDelay = 1000;
  let reconnectTimer = null;

  const messagesCount = computed(() => messages.value.length);

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
      reconnectDelay = 1000;
    };

    ws.onmessage = (event) => {
      if (socket.value !== ws) return;
      try {
        const data = JSON.parse(event.data);
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
        if (data.id && messages.value.some((m) => m.id === data.id)) return;
        messages.value.push(data);
      } catch {
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
        reconnectDelay = Math.min(reconnectDelay * 2, 30000);
      }
    };
  }

  function disconnect() {
    intentionalClose = true;
    clearTimeout(reconnectTimer);
    if (!socket.value) return;
    const ws = socket.value;
    socket.value = null;
    isConnected.value = false;
    ws.close();
  }

  function sendMessage(text, isPrivate = false) {
    if (!socket.value || !isConnected.value) return;
    socket.value.send(JSON.stringify({ text, private: isPrivate }));
  }

  function deleteMessage(id) {
    if (!socket.value || !isConnected.value) return;
    socket.value.send(JSON.stringify({ action: "delete", id }));
  }

  function clearMessages() {
    messages.value = [];
  }

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
