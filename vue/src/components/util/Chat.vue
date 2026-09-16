<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import Button from "@/components/input/Button.vue";
import ToggleButton from "@/components/input/ToggleButton.vue";
import { useMessagesStore } from "@/stores/messages";
import { useAuthStore } from "@/stores/auth";
import Header from "@/components/text/Header.vue";
import Link from "@/components/text/Link.vue";

const messagesStore = useMessagesStore();
const authStore = useAuthStore();
const messages = computed(() => messagesStore.messages);
const messageInput = ref("");
const messagesContainer = ref(null);
const messagesInner = ref(null);
const fileInput = ref(null);
const isAdmin = computed(() => !!authStore.user.admin);
const sendPrivate = ref(false);

const isNearBottom = ref(true);
const SCROLL_THRESHOLD = 100;
let resizeObserver = null;

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  }
}

function scrollToBottomIfNear() {
  if (isNearBottom.value) {
    scrollToBottom();
  }
}

function onScroll() {
  if (!messagesContainer.value) return;
  const { scrollHeight, scrollTop, clientHeight } = messagesContainer.value;
  isNearBottom.value =
    scrollHeight - scrollTop - clientHeight < SCROLL_THRESHOLD;
}

function goToBottom() {
  isNearBottom.value = true;
  scrollToBottom();
}

watch(
  () => messages.value.length,
  () => {
    nextTick(scrollToBottomIfNear);
  },
);

function sendMessage() {
  const text = messageInput.value.trim();
  if (!text) return;
  isNearBottom.value = true;
  messagesStore.sendMessage(text, isAdmin.value && sendPrivate.value);
  messageInput.value = "";
}

function deleteMessage(id) {
  messagesStore.deleteMessage(id);
}

// Admin status is fixed when the socket connects (from the auth cookies), so
// reconnect after a login/logout to pick up or drop private messages.
watch(isAdmin, () => {
  if (!messagesStore.isConnected) return;
  messagesStore.disconnect();
  messagesStore.connect();
});

async function onFileSelected(e) {
  const file = e.target.files[0];
  if (!file) return;
  isNearBottom.value = true;
  await messagesStore.uploadAndSendFile(
    file,
    isAdmin.value && sendPrivate.value,
  );
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

  if (messagesContainer.value) {
    messagesContainer.value.addEventListener("scroll", onScroll, {
      passive: true,
    });
  }

  if (messagesInner.value) {
    resizeObserver = new ResizeObserver(scrollToBottomIfNear);
    resizeObserver.observe(messagesInner.value);
  }

  scrollToBottom();
});

onUnmounted(() => {
  messagesStore.disconnect();

  if (messagesContainer.value) {
    messagesContainer.value.removeEventListener("scroll", onScroll);
  }

  if (resizeObserver) {
    resizeObserver.disconnect();
    resizeObserver = null;
  }
});
</script>

<template>
  <div class="chat-root flex-col flex min-h-0">
    <Header>Chat</Header>
    <div
      ref="messagesContainer"
      class="flex flex-col flex-1 min-h-0 overflow-y-auto overflow-x-hidden p-2 min-w-0"
    >
      <div ref="messagesInner">
        <p
          v-for="message in messages"
          :key="message.id"
          class="break-words min-w-0 w-full"
          :class="{ 'text-secondary italic': message.private }"
        >
          <button
            v-if="isAdmin && message.id"
            type="button"
            class="text-tertiary hover:text-primary cursor-pointer mr-1"
            aria-label="Delete message"
            title="Delete message"
            @click="deleteMessage(message.id)"
          >
            ×
          </button>
          <span
            v-if="message.private"
            class="text-tertiary"
            title="Private: only admins can see this"
            aria-label="Private message"
            >🔒</span
          >
          <span class="text-tertiary">{{ message.authorId }}:</span>
          <template
            v-for="(part, i) in parseMessageParts(message.text || '')"
            :key="i"
          >
            <Link
              v-if="part.type === 'link'"
              bare
              :href="part.value"
              target="_blank"
              class="text-primary underline break-all"
              >{{ part.value }}</Link
            >
            <span v-else>{{ part.value }}</span>
          </template>
          <template v-if="message.fileUrl && isSafeFileUrl(message.fileUrl)">
            <img
              v-if="isImageUrl(message.fileUrl)"
              :src="message.fileUrl"
              alt="Uploaded image"
              loading="lazy"
              class="w-full max-w-full rounded block"
            />
            <video
              v-else-if="isVideoUrl(message.fileUrl)"
              :src="message.fileUrl"
              controls
              preload="none"
              class="w-full max-w-full max-h-48 rounded block"
            />
            <Link
              v-else
              bare
              :href="message.fileUrl"
              target="_blank"
              class="underline break-all"
              >{{ message.fileUrl.split("/").pop() }}</Link
            >
          </template>
        </p>
      </div>
    </div>
    <div>
      <input
        v-model="messageInput"
        @keyup.enter="sendMessage"
        aria-label="Chat message"
      />
      <input
        ref="fileInput"
        type="file"
        class="hidden"
        @change="onFileSelected"
      />
      <label
        v-if="isAdmin"
        class="flex items-center gap-2 py-1 text-sm text-secondary cursor-pointer"
      >
        <ToggleButton v-model="sendPrivate" />
        <span>Private (admins only)</span>
      </label>
      <div class="flex gap-2">
        <Button class="flex-1" @click="sendMessage">Send</Button>
        <Button
          v-if="authStore.user.admin"
          class="flex-1"
          @click="fileInput.click()"
          >Attach</Button
        >
        <Button v-if="!isNearBottom" class="flex-1" @click="goToBottom"
          >Bottom</Button
        >
      </div>
    </div>
  </div>
</template>

<style scoped>
@media (max-width: 850px) {
  .chat-root {
    max-height: none;
    height: 100%;
  }
}
</style>
