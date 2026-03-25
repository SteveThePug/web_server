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
    if (socket.value && isConnected.value) return;
    intentionalClose = false;

    socket.value = new WebSocket(getWebSocketURL());

    socket.value.onopen = () => {
      isConnected.value = true;
      lastError.value = null;
      reconnectDelay = 1000;
      messages.value = [];
    };

    socket.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        messages.value.push(data);
      } catch {
        messages.value.push({ text: event.data });
      }
    };

    socket.value.onerror = (error) => {
      lastError.value = error;
    };

    socket.value.onclose = () => {
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
    socket.value.close();
    socket.value = null;
    isConnected.value = false;
  }

  function sendMessage(text) {
    if (!socket.value || !isConnected.value) return;
    socket.value.send(JSON.stringify({ text }));
  }

  function clearMessages() {
    messages.value = [];
  }

  async function uploadAndSendFile(file) {
    try {
      const formData = new FormData();
      formData.append("file", file);
      const res = await axios.post("/api/messages/upload", formData);
      const { url } = res.data;
      if (!socket.value || !isConnected.value) return;
      socket.value.send(JSON.stringify({ text: "", fileUrl: url }));
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
    clearMessages,
    uploadAndSendFile,
  };
});
