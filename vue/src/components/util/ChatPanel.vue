<script setup>
/**
 * One chat channel: message list, composer, file attach and admin delete.
 * Drives stores/messages.js. Not a widget itself — Chat.vue shows the public
 * channel and PrivateChat.vue the admins-only one.
 *
 * The socket carries both channels in one stream (the server only sends
 * private messages to admin connections), so `isPrivate` picks which half of
 * `messages` to show and also marks what you send — you cannot post to a
 * channel you are not looking at.
 *
 * Holds the socket while mounted — connect() on mount, disconnect() on
 * unmount; the store counts them, so two panels share one socket and it
 * exists only while a chat is on screen.
 */
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from "vue";
import Button from "@/components/input/Button.vue";
import { useMessagesStore } from "@/stores/messages";
import { useAuthStore } from "@/stores/auth";
import Header from "@/components/text/Header.vue";
import Link from "@/components/text/Link.vue";
import { ACCEPT, isUploadUrl, kindOf } from "@/uploads";

const messagesStore = useMessagesStore();
const authStore = useAuthStore();
const messageInput = ref("");
const messagesContainer = ref(null);
const messagesInner = ref(null);
const fileInput = ref(null);
const isAdmin = computed(() => !!authStore.user.admin);

const props = defineProps({
  // Only PrivateChat.vue sets this, and it mounts the panel for admins only.
  isPrivate: { type: Boolean, default: false },
});

// The private panel is the only place private messages are shown, so the
// public one stays public even for an admin.
const messages = computed(() =>
  messagesStore.messages.filter((m) => !!m.private === props.isPrivate),
);

// Autoscroll only when the reader is already at the bottom: scrolling up to
// read history must not be yanked back down by an incoming message.
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

// nextTick: scrollHeight is only correct once Vue has patched the new row in.
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
  messagesStore.sendMessage(text, props.isPrivate);
  messageInput.value = "";
}

function deleteMessage(id) {
  messagesStore.deleteMessage(id);
}

async function onFileSelected(e) {
  const file = e.target.files[0];
  if (!file) return;
  isNearBottom.value = true;
  // The private flag both routes the upload to /uploads/private/ and marks the
  // message itself private; the two must agree.
  await messagesStore.uploadAndSendFile(file, props.isPrivate);
  fileInput.value.value = "";
}

// Attachments are rendered only if isUploadUrl() accepts them: the fileUrl
// arrives over the socket and is otherwise attacker-controlled.

// NOTE: a module-level /g regex carries `lastIndex` between calls. It is safe
// here only because parseMessageParts() below always runs exec() to completion
// (exec returns null, resetting lastIndex) before returning. An early return
// from that loop would make the next message parse from the wrong offset.
const urlRegex = /(https?:\/\/[^\s<]+)/g;

/** Split message text into alternating {type:'text'|'link', value} parts so
 *  URLs can be rendered as anchors without using v-html. */
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

  // Images and videos load after their row is inserted and change its height,
  // which the length watcher above has already missed. Observing the inner
  // wrapper catches those late reflows.
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
    <Header>{{ isPrivate ? "🔒 Private Chat" : "Chat" }}</Header>
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
          <template v-if="isUploadUrl(message.fileUrl)">
            <img
              v-if="kindOf(message.fileUrl) === 'image'"
              :src="message.fileUrl"
              alt="Uploaded image"
              loading="lazy"
              class="w-full max-w-full rounded block"
            />
            <video
              v-else-if="kindOf(message.fileUrl) === 'video'"
              :src="message.fileUrl"
              controls
              preload="none"
              class="w-full max-w-full max-h-48 rounded block"
            />
            <audio
              v-else-if="kindOf(message.fileUrl) === 'audio'"
              :src="message.fileUrl"
              controls
              preload="none"
              class="w-full max-w-full block"
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
        :aria-label="isPrivate ? 'Private chat message' : 'Chat message'"
        :placeholder="isPrivate ? 'Admins only' : ''"
      />
      <input
        ref="fileInput"
        type="file"
        :accept="ACCEPT.any"
        class="hidden"
        @change="onFileSelected"
      />
      <div class="flex gap-2">
        <Button class="flex-1" @click="sendMessage">Send</Button>
        <Button v-if="isAdmin" class="flex-1" @click="fileInput.click()"
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
