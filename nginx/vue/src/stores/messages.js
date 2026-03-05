import { defineStore } from "pinia";
import { ref, computed } from "vue";

function getWebSocketURL() {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.host}/api/ws`;
}

export const useMessagesStore = defineStore("messages", () => {
  const socket = ref(null);
  const messages = ref([]);
  const isConnected = ref(false);
  const lastError = ref(null);

  const messagesCount = computed(() => messages.value.length);

  function connect() {
    if (socket.value && isConnected.value) return;

    socket.value = new WebSocket(getWebSocketURL());

    socket.value.onopen = () => {
      isConnected.value = true;
      lastError.value = null;
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
    };
  }

  function disconnect() {
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

  return {
    messages,
    isConnected,
    lastError,

    messagesCount,

    connect,
    disconnect,
    sendMessage,
    clearMessages,
  };
});
