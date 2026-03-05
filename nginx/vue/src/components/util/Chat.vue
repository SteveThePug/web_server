<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import Button from "@/components/input/Button.vue";
import { useMessagesStore } from "@/stores/messages";

const messagesStore = useMessagesStore();
const messages = computed(() => messagesStore.messages);
const messageInput = ref("");

function sendMessage() {
    const text = messageInput.value.trim();
    if (!text) return;
    messagesStore.sendMessage(text);
    messageInput.value = "";
}

onMounted(() => {
    messagesStore.connect();
});
onUnmounted(() => {
    messagesStore.disconnect();
});
</script>

<template>
    <div class="flex flex-col place-content-around h-full">
        <div class="flex flex-col">
            <p v-for="message in messages" :key="message.id">
                {{ message.text }}
            </p>
        </div>
        <input v-model="messageInput" @keyup.enter="sendMessage" />
        <Button @click="sendMessage">Send</Button>
    </div>
</template>
