<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import Button from "@/components/input/Button.vue";
import { useMessagesStore } from "@/stores/messages";
import Header from "@/components/text/Header.vue";

const messagesStore = useMessagesStore();
const messages = computed(() => messagesStore.messages);
const messageInput = ref("");
const messagesContainer = ref(null);

function scrollToBottom() {
    nextTick(() => {
        if (messagesContainer.value) {
            messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
        }
    });
}

watch(messages, scrollToBottom, { deep: true });

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
    <div class="flex flex-col">
        <Header>Chat</Header>
        <div ref="messagesContainer" class="flex flex-col flex-1 overflow-y-auto">
            <p v-for="message in messages" :key="message.id">
                <span class="font-bold">User {{ message.authorId }}:</span> {{ message.text }}
            </p>
        </div>
        <input v-model="messageInput" @keyup.enter="sendMessage" />
        <Button @click="sendMessage">Send</Button>
    </div>
</template>
