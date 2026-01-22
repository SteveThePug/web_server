<script setup>
import { onMounted, onUnmounted, ref } from "vue";
// Connect to websocket
const url = "/api/ws";

const socket = ref(null);
const messages = ref([]);
const messageInput = ref("");

onMounted(() => {
    socket.value = new WebSocket(url);

    socket.value.addEventListener("message", (event) => {
        const message = JSON.parse(event.data);
        messages.value.push(message);
    });
});

function sendMessage() {
    socket.value.send(JSON.stringify({ content: messageInput.value }));
    messageInput.value = "";
}

onUnmounted(() => {
    socket.value?.close();
});
</script>

<template>
    <div>
        <div class="flex-col">
            <p v-for="message in messages" :key="message.id">
                {{ message.content }}
            </p>
        </div>
        <div class="flex-row">
            <input v-model="messageInput" @keyup.enter="sendMessage" />
            <button @click="sendMessage">Send</button>
        </div>
    </div>
</template>
