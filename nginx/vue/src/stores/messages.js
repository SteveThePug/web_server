import { defineStore } from "pinia";
import { ref, computed } from "vue";

const URL = "/api/ws";

const message_template = {
  id: 0,
  content: "Yo",
};

export const useMessagesStore = defineStore("messages", () => {
  const socket = ref(null);
  const messages = ref([message_template]);
  const isConnected = ref(false);
  const lastError = ref(null);

  const messagesCount = computed(() => messages.value.length);

  function connect() {
    if (socket.value && isConnected.value) return;

    socket.value = new WebSocket(URL);

    socket.value.onopen = () => {
      isConnected.value = true;
      lastError.value = null;
    };

    socket.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        messages.value.push(data);
      } catch {
        // fallback if server sends plain text
        messages.value.push(event.data);
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

  function sendMessage(payload) {
    if (!socket.value || !isConnected.value) return;
    socket.value.send(JSON.stringify(payload));
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
