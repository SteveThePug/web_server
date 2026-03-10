<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import Button from "@/components/input/Button.vue";
import { useMessagesStore } from "@/stores/messages";
import { useAuthStore } from "@/stores/auth";
import Header from "@/components/text/Header.vue";

const messagesStore = useMessagesStore();
const authStore = useAuthStore();
const messages = computed(() => messagesStore.messages);
const messageInput = ref("");
const messagesContainer = ref(null);
const fileInput = ref(null);

function scrollToBottom() {
    nextTick(() => {
        if (messagesContainer.value) {
            messagesContainer.value.scrollTop =
                messagesContainer.value.scrollHeight;
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

async function onFileSelected(e) {
    const file = e.target.files[0];
    if (!file) return;
    await messagesStore.uploadAndSendFile(file);
    fileInput.value.value = "";
}

function isImageUrl(url) {
    return /\.(jpg|jpeg|png|gif|webp)$/i.test(url);
}

function isVideoUrl(url) {
    return /\.(mp4|webm|ogg|mov)$/i.test(url);
}

function isSafeFileUrl(url) {
    return typeof url === "string" && url.startsWith("/uploads/");
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
        <div
            ref="messagesContainer"
            class="flex flex-col flex-1 overflow-y-auto p-2"
        >
            <p v-for="message in messages" :key="message.id">
                <span class="text-tertiary">{{ message.authorId }}:</span>
                {{ message.text }}
                <template
                    v-if="message.fileUrl && isSafeFileUrl(message.fileUrl)"
                >
                    <img
                        v-if="isImageUrl(message.fileUrl)"
                        :src="message.fileUrl"
                        class="max-w-xs max-h-48 rounded"
                    />
                    <video
                        v-else-if="isVideoUrl(message.fileUrl)"
                        :src="message.fileUrl"
                        controls
                        class="max-w-xs max-h-48 rounded"
                    />
                    <a
                        v-else
                        :href="message.fileUrl"
                        target="_blank"
                        class="underline"
                        >{{ message.fileUrl.split("/").pop() }}</a
                    >
                </template>
            </p>
        </div>
        <input v-model="messageInput" @keyup.enter="sendMessage" />
        <input
            ref="fileInput"
            type="file"
            class="hidden"
            @change="onFileSelected"
        />
        <div class="flex justify-between">
            <Button @click="sendMessage">Send</Button>
            <Button v-if="authStore.user.admin" @click="fileInput.click()"
                >Attach</Button
            >
            <div class="flex gap-2">
                <Button @click="fileInput.click()">Attach</Button>
                <Button @click="scrollToBottom">To Bottom</Button>
            </div>
        </div>
    </div>
</template>
