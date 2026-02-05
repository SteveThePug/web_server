<script setup>
import Button from "@/components/input/Button.vue";
import { useMessagesStore } from "@/stores/messages";

const messagesStore = useMessagesStore();
const messages = computed(() => messagesStore.messages);

onMounted(() => {
    messagesStore.connect();
});
onUnmounted(() => {
    messagesStore.disconnect();
});
</script>

<template>
    <div>
        <div class="flex flex-col">
            <p v-for="message in messages" :key="message.id">
                {{ message.content }}
            </p>
        </div>
        <div class="flex flex-row">
            <input v-model="messageInput" @keyup.enter="sendMessage" />
            <Button @click="sendMessage">Send</Button>
        </div>
    </div>
</template>
