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

const urlRegex = /(https?:\/\/[^\s<]+)/g;

function parseMessageParts(text) {
    const parts = [];
    let lastIndex = 0;
    let match;
    while ((match = urlRegex.exec(text)) !== null) {
        if (match.index > lastIndex) {
            parts.push({
                type: "text",
                value: text.slice(lastIndex, match.index),
            });
        }
        parts.push({ type: "link", value: match[1] });
        lastIndex = urlRegex.lastIndex;
    }
    if (lastIndex < text.length) {
        parts.push({ type: "text", value: text.slice(lastIndex) });
    }
    return parts;
}

onMounted(() => {
    messagesStore.connect();
});
onUnmounted(() => {
    messagesStore.disconnect();
});
</script>

<template>
    <div>
        <Header>Chat</Header>
        <div class="flex-col flex">
            <div
                ref="messagesContainer"
                class="flex flex-col overflow-y-auto p-2"
            >
                <p v-for="message in messages" :key="message.id">
                    <span class="text-tertiary">{{ message.authorId }}:</span>
                    <template
                        v-for="(part, i) in parseMessageParts(
                            message.text || '',
                        )"
                        :key="i"
                    >
                        <a
                            v-if="part.type === 'link'"
                            :href="part.value"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="text-primary underline"
                            >{{ part.value }}</a
                        >
                        <span v-else>{{ part.value }}</span>
                    </template>
                    <template
                        v-if="message.fileUrl && isSafeFileUrl(message.fileUrl)"
                    >
                        <img
                            v-if="isImageUrl(message.fileUrl)"
                            :src="message.fileUrl"
                            class="max-w-xs max-h-48 rounded"
                            @load="scrollToBottom"
                        />
                        <video
                            v-else-if="isVideoUrl(message.fileUrl)"
                            :src="message.fileUrl"
                            controls
                            class="max-w-xs max-h-48 rounded"
                            @loadedmetadata="scrollToBottom"
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
            <div>
                <input v-model="messageInput" @keyup.enter="sendMessage" />
                <input
                    ref="fileInput"
                    type="file"
                    class="hidden"
                    @change="onFileSelected"
                />
                <div class="flex gap-2">
                    <Button class="flex-1" @click="sendMessage">Send</Button>
                    <Button
                        v-if="authStore.user.admin"
                        class="flex-1"
                        @click="fileInput.click()"
                        >Attach</Button
                    >
                    <Button class="flex-1" @click="scrollToBottom"
                        >Bottom</Button
                    >
                </div>
            </div>
        </div>
    </div>
</template>
